package main

import (
	"math"
	"math/rand"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
)

// 初始数值
const (
	BombMaxTime = 4 // 炸弹的持续时间

	HurtScore = 1 // 造成伤害得分
	KillScore = 3 // 击杀得分

	RoleInitBombPower = 1  // 初始炸弹威力
	RoleInitSpeed     = 1  // 初始移动速度
	RoleInitHp        = 10 // 玩家初始血量
	RoleInitPosX      = 11 // 玩家初始位置x
	RoleInitPosY      = 15 // 玩家初始位置y
)

type ScenePlayer struct {
	id      uint64           // 玩家id
	name    string           // 玩家昵称
	key     string           // 登录token
	score   uint32           // 分数
	curPos  *common.Position // 当前位置
	nextPos *common.Position // 下一个位置
	hp      int32            // 生命值
	curbomb uint32           // 当前已经放置的炸弹数量
	maxbomb uint32           // 能放置的最大炸弹数量
	power   uint32           // 炸弹威力
	speed   float64          // 移动速度

	self  *PlayerTask // 用于通信
	scene *Scene      // 场景指针

	isMove    bool
	deathTime time.Time // 玩家死亡时间（用于计算存活时间）

	// AOI减少视野
	watcher   map[uint64]*ScenePlayer // 观察者集合（此时所有关注当前角色的对象）
	beWatcher map[uint64]*ScenePlayer // 被观察者集合（此时当前角色关注的所有对象）
	vision    float64                 // 视野
	// AOI九宫格
	aoiIdx int
}

func NewScenePlayer(player *PlayerTask, scene *Scene) *ScenePlayer {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	_ = r
	s := &ScenePlayer{
		id:      player.id,
		name:    player.name,
		key:     player.key,
		score:   0,
		curPos:  &common.Position{X: RoleInitPosX, Y: float64(RoleInitPosY - (player.id%100)/10)},
		nextPos: &common.Position{X: RoleInitPosX, Y: float64(RoleInitPosY - (player.id%100)/10)},
		hp:      RoleInitHp,
		curbomb: 0,
		maxbomb: 5,
		power:   RoleInitBombPower,
		speed:   RoleInitSpeed,

		self:   player,
		scene:  scene,
		isMove: false,

		watcher:   make(map[uint64]*ScenePlayer),
		beWatcher: make(map[uint64]*ScenePlayer),
		vision:    5.0,
	}
	return s
}

// ----------------------玩家角色事件-------------------------- //

// 放置炸弹
func (this *ScenePlayer) PutBomb(msg *usercmd.MsgPutBomb) bool {
	// 达到最大炸弹数
	if atomic.LoadUint32(&this.curbomb) == this.maxbomb {
		return false
	}
	// 当前位置是否已经存在炸弹
	x, y := this.GetCurrentGrid()
	if this.scene.gameMap.MapArray[x][y] == GridType_Bomb {
		glog.Infof("[%v 放置炸弹] 位置{%v, %v}, 当前位置已存在炸弹",
			this.name, x, y)
		return false
	}

	bomb := NewBomb(this)
	this.scene.AddBomb(bomb)
	go bomb.CountDown()

	atomic.StoreUint32(&this.curbomb, this.curbomb+1)
	glog.Infof("[%v 放置炸弹] 炸弹位置{%v, %v}, 已放炸弹:%v, 剩余炸弹:%v",
		bomb.owner.name, bomb.pos.X, bomb.pos.Y, this.curbomb, this.maxbomb-this.curbomb)
	return true
}

// 移动
func (this *ScenePlayer) Move(msg *usercmd.MsgMove) {
	// TODO BUG:invalid memory address or nil pointer dereference
	if this == nil {
		return
	}
	this.isMove = true

	this.CaculateNext(msg.Way)     // 计算下一个位置
	this.BorderCheck(this.nextPos) // 边界检查

	this.curPos = this.nextPos

	// TODO aoi（待测试）
	this.AoiMove()
	// this.MoveAoiGrid()

	glog.Infof("[%v 移动]当前位置为 x:%v, y:%v",
		this.name, this.nextPos.X, this.nextPos.Y)
}

// 根据角度移动
// func (this *ScenePlayer) CaculateNext(direct float64) {
// 	this.nextPos.X = this.curPos.X + math.Sin(direct*math.Pi/180)*float64(this.speed)
// 	this.nextPos.Y = this.curPos.Y + math.Cos(direct*math.Pi/180)*float64(this.speed)
// }

// TODO 上下左右移动
func (this *ScenePlayer) CaculateNext(way int32) {
	x, y := this.GetCurrentGrid()
	// glog.Errorf("[GetCurrentGrid] x:%v, y:%v", x, y)
	// 上1下2左3右4
	switch common.MoveWay(way) {
	case common.MoveWay_Up:
		if this.CanPass(x, y+1) {
			this.nextPos.X = float64(x) // 如果在格子边缘，自动调整到格子中央
			this.nextPos.Y = this.curPos.Y + this.speed
		}
		break
	case common.MoveWay_Down:
		if this.CanPass(x, y-1) {
			this.nextPos.X = float64(x)
			this.nextPos.Y = this.curPos.Y - this.speed
		}
		break
	case common.MoveWay_Left:
		if this.CanPass(x-1, y) {
			this.nextPos.X = this.curPos.X - this.speed
			this.nextPos.Y = float64(y)
		}
		break
	case common.MoveWay_Right:
		if this.CanPass(x+1, y) {
			this.nextPos.X = this.curPos.X + this.speed
			this.nextPos.Y = float64(y)
		}
		break
	default:
	}
}

// 地图边界检查
func (this *ScenePlayer) BorderCheck(pos *common.Position) {
	this.scene.BorderCheck(pos)
}

// 该格子是否可以通过
func (this *ScenePlayer) CanPass(x, y uint32) bool {
	return this.scene.CanPass(x, y)
}

// 判断该玩家当前属于哪一个格子
func (this *ScenePlayer) GetCurrentGrid() (uint32, uint32) {
	return uint32(common.Round(this.curPos.X)),
		uint32(common.Round(this.curPos.Y))
}

// 收到伤害
func (this *ScenePlayer) BeHurt(attacker *ScenePlayer) {
	if atomic.StoreInt32(&this.hp, this.hp-1); this.hp <= 0 {
		// TODO 玩家角色死亡
		glog.Infoln("[玩家死亡] username:", this.name)
		info := &usercmd.RetRoleDeath{
			KillName: attacker.name,
			KillId:   attacker.id,
			LiveTime: 0,
			Score:    this.score,
		}
		// 向客户端发送玩家死亡信息
		this.self.SendCmd(usercmd.MsgTypeCmd_Death, info)
		// 玩家死亡房间/场景处理
		this.Death()
	}
	glog.Infof("[%v收到%v的炸弹的伤害] 当前血量hp:%v，%v当前得分:%v",
		this.name, attacker.name, this.hp, attacker.name, attacker.score)
}

// 玩家造成伤害或击杀，增加得分
func (this *ScenePlayer) AddScore(x uint32) {
	atomic.StoreUint32(&this.score, this.score+x)
	atomic.StoreUint32(&this.self.score, this.self.score+x)
}

// ---------------------------------------------------------- //

// 发送场景同步信息
func (this *ScenePlayer) SendSceneMessage() {
	ret := &usercmd.RetUpdateSceneMsg{}
	// 场景内所有的玩家信息
	ret.Id = this.id
	for _, player := range this.scene.players {
		ret.Players = append(ret.Players, &usercmd.ScenePlayer{
			Id:      player.id,
			BombNum: player.curbomb,
			Power:   player.power,
			Speed:   float32(player.speed),
			State:   uint32(player.hp),
			X:       float32(player.curPos.X),
			Y:       float32(player.curPos.Y),
			IsMove:  player.isMove,
			Score:   this.score,
		})
	}
	// 场景内所有的炸弹信息
	for _, bomb := range this.scene.BombMap {
		ret.Bombs = append(ret.Bombs, &usercmd.MsgBomb{
			Id:         bomb.id,
			Own:        bomb.owner.id,
			X:          int32(bomb.pos.X),
			Y:          int32(bomb.pos.Y),
			CreateTime: 0,
		})
	}

	this.self.SendCmd(usercmd.MsgTypeCmd_SceneSync, ret)
}

func (this *ScenePlayer) Update() {

}

// 玩家死亡处理
func (this *ScenePlayer) Death() {
	this.scene.DelPlayer(this)             // 场景中删除
	this.self.room.RemovePlayer(this.self) // 把玩家从房间中删除

}

// 添加玩家数据到场景同步信息
func (this *ScenePlayer) AddAllPlayerInfoToMessage(msg common.Message) {
	// 类型断言
	if ret, ok := msg.(*usercmd.RetUpdateSceneMsg); ok {

		for _, player := range this.scene.players {
			ret.Players = append(ret.Players, &usercmd.ScenePlayer{
				Id:      player.id,
				BombNum: player.curbomb,
				Power:   player.power,
				Speed:   float32(player.speed),
				State:   uint32(player.hp),
				X:       float32(player.curPos.X),
				Y:       float32(player.curPos.Y),
				Score:   this.score,
				IsMove:  player.isMove,
			})
		}
	}
}

// 计算两个场景内玩家的距离
func (this *ScenePlayer) DistanceTo(player *ScenePlayer) float64 {
	return math.Sqrt(math.Pow((this.curPos.X-player.curPos.X), 2) +
		math.Pow((this.curPos.Y-player.curPos.Y), 2))
}

// --------------------------------AOI（减少视野，全局搜索）-------------------------------- //
// AoiEnter 玩家进入场景（向指定的玩家列表发送玩家进入信息）
func (this *ScenePlayer) AoiEnter() {
	for _, player := range this.scene.players {
		// 玩家在我的视野内
		if this.DistanceTo(player) <= this.vision {
			player.EnterVision(this)
		}
		// 我在玩家的视野内
		if player.DistanceTo(this) <= player.vision {
			this.EnterVision(player)
		}
	}
}

// AoiMove 玩家在场景中移动
func (this *ScenePlayer) AoiMove() {
	// 遍历场景内的所有对象
	for _, player := range this.scene.players {
		// 如果该玩家原来在【我的被观察者集合】中，并且现在的距离已经大于【我的视野】(我原来看得到他，现在看不到)
		if _, ok := this.beWatcher[player.id]; ok && this.DistanceTo(player) > this.vision {
			player.LeaveVision(this)
		}
		// 如果该玩家原来在【我的观察者集合】中，并且现在的距离已经大于【它的视野】（他原来看的到我，现在看不到）
		if _, ok := this.watcher[player.id]; ok && this.DistanceTo(player) > player.vision {
			this.LeaveVision(player)
		}

		// 如果该玩家原来不在【我的被观察者集合】中，并且现在的距离已经小于【我的视野】（我原来看不到他，现在看得到）
		if _, ok := this.beWatcher[player.id]; !ok && this.DistanceTo(player) <= this.vision {
			player.EnterVision(this)
		}
		// 如果该玩家原来不在【我的观察者集合】中，并且现在的距离已经小于【它的视野】（他原来看不到我，现在看得到）
		if _, ok := this.watcher[player.id]; !ok && this.DistanceTo(player) <= player.vision {
			this.EnterVision(player)
		}
	}
}

// AoiLeave 玩家从场景中离开
func (this *ScenePlayer) AoiLeave() {
	// 遍历我的观察者集合，向他们发送Leave(我)事件，此时我从对象的被观察者集合中删除，同时对象也会从我的观察者集合删除
	for _, player := range this.watcher {
		player.LeaveVision(this)
	}
	// 遍历我的被观察者集合，将我从这些对象的观察者集合中删除，将我的被观察者集合清空
	for _, player := range this.beWatcher {
		this.LeaveVision(player)
	}
	this.beWatcher = make(map[uint64]*ScenePlayer, 0)
}

// 当前玩家离开另一个玩家p的视野，此时p会从我的被观察者集合中删除，同时我会从p的观察者集合中删除
func (this *ScenePlayer) LeaveVision(player *ScenePlayer) {
	delete(this.beWatcher, player.id)
	delete(player.watcher, this.id)
}

// 当前玩家进入另一个玩家p的视野，此时p会从我的观察者集合中添加，同时我会从p的被观察者集合中添加
func (this *ScenePlayer) EnterVision(player *ScenePlayer) {
	this.watcher[player.id] = player
	player.beWatcher[this.id] = this
}

// -------------------------------AOI九宫格------------------------------- //

// GetBroadCastPlayerList 获取九宫格内的所有玩家（需要广播信息的玩家列表）
func (this *ScenePlayer) GetBroadCastPlayerList() []*ScenePlayer {
	// 获取当前玩家所在aoi九宫格的所有9个格子index
	idx := int(CaculateAoiGrid(this.GetCurrentGrid()))
	gridIdxs := make([]int, 0)
	size := int(MaxGridNumberX / AoiGridSize)
	gridIdxs = append(gridIdxs, idx+size-1, idx+size, idx+size+1,
		idx-1, idx, idx+1,
		idx-size-1, idx-size, idx-size+1)
	// 遍历aoi九宫格
	res := make([]*ScenePlayer, 0)
	for _, index := range gridIdxs {
		for _, pl := range this.scene.aoiVector[index].players {
			res = append(res, pl)
		}
	}
	return res
}

// EnterAoiGrid 玩家第一次进入aoi宫格
func (this *ScenePlayer) EnterAoiGrid() {
	players := this.GetBroadCastPlayerList()
	for _, pl := range players {
		// 当前玩家看到其他玩家
		this.beWatcher[pl.id] = pl
		pl.watcher[this.id] = this
		// 其他玩家看到当前玩家
		this.watcher[pl.id] = pl
		pl.beWatcher[this.id] = this
	}
}

// MoveAoiGrid 玩家在aoi宫格中移动
func (this *ScenePlayer) MoveAoiGrid() {
	// 计算是否进入了新的aoi九宫格
	newAoiIdx := int(CaculateAoiGrid(this.GetCurrentGrid()))
	if newAoiIdx != this.aoiIdx {
		size := int(MaxGridNumberX / AoiGridSize)
		var newIdxList, oldIdxList [3]int
		// 进入新的九宫格
		switch newAoiIdx {
		case this.aoiIdx + size: // 上
			newIdxList = [3]int{newAoiIdx + size - 1, newAoiIdx + size, newAoiIdx + size + 1}
			oldIdxList = [3]int{this.aoiIdx - size - 1, this.aoiIdx - size, this.aoiIdx - size + 1}
		case this.aoiIdx - size: // 下
			newIdxList = [3]int{newAoiIdx - size - 1, newAoiIdx - size, newAoiIdx - size + 1}
			oldIdxList = [3]int{this.aoiIdx + size - 1, this.aoiIdx + size, this.aoiIdx + size + 1}
		case this.aoiIdx - 1: // 左
			newIdxList = [3]int{newAoiIdx + size - 1, newAoiIdx - 1, newAoiIdx - size - 1}
			oldIdxList = [3]int{this.aoiIdx + size + 1, this.aoiIdx + 1, this.aoiIdx - size + 1}
		case this.aoiIdx + 1: // 右
			newIdxList = [3]int{newAoiIdx + size + 1, newAoiIdx + 1, newAoiIdx - size + 1}
			oldIdxList = [3]int{this.aoiIdx + size - 1, this.aoiIdx - 1, this.aoiIdx - size - 1}
		}
		// 对新的三个宫格发送enter信息
		for _, idx := range newIdxList {
			for _, player := range this.scene.GetAoiPlayersByIdx(idx) {
				this.EnterVision(player)
				player.EnterVision(this)
			}
		}
		// 对旧的三个宫格发送leave信息
		for _, idx := range oldIdxList {
			for _, player := range this.scene.GetAoiPlayersByIdx(idx) {
				this.LeaveVision(player)
				player.LeaveVision(this)
			}
		}
		// 旧宫格中移除，新宫格中添加
		this.scene.AoiAddPlayer(newAoiIdx, this)
		this.scene.AoiDelPlayer(this.aoiIdx, this)
		// 更新当前的宫格坐标
		this.aoiIdx = newAoiIdx

	} else {

	}
}

// LeaveAoiGrid 玩家离开场景
func (this *ScenePlayer) LeaveAoiGrid() {
	players := this.GetBroadCastPlayerList()
	for _, pl := range players {
		// 当前玩家看不到其他玩家
		delete(this.beWatcher, pl.id)
		delete(pl.watcher, this.id)
		// 其他玩家看不到当前玩家
		delete(this.watcher, pl.id)
		delete(pl.beWatcher, this.id)
	}
}
