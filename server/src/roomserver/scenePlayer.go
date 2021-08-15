package main

import (
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
	RoleInitHp        = 3  // 玩家初始血量
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
