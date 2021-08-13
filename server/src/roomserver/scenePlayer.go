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

	RoleInitBombPower = 1 // 初始炸弹威力
	RoleInitSpeed     = 1 // 初始移动速度
	RoleInitHp        = 3 // 玩家初始血量
	RoleInitPosX      = 7 // 玩家初始位置x
	RoleInitPosY      = 7 // 玩家初始位置y
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

	isMove bool
}

func NewScenePlayer(player *PlayerTask, scene *Scene) *ScenePlayer {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	_ = r
	s := &ScenePlayer{
		id:      player.id,
		name:    player.name,
		key:     player.key,
		score:   0,
		curPos:  &common.Position{X: RoleInitPosX, Y: RoleInitPosY},
		nextPos: &common.Position{X: RoleInitPosX, Y: RoleInitPosY},
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
	if pos.X < 0 {
		pos.X = 0
	} else if w := float64(this.scene.sceneWidth - 1); pos.X >= w {
		pos.X = w
	}
	if pos.Y < 0 {
		pos.Y = 0
	} else if h := float64(this.scene.sceneHeight - 1); pos.Y >= h {
		pos.Y = h
	}
}

// 该格子是否可以通过
func (this *ScenePlayer) CanPass(x, y uint32) bool {
	if x >= this.scene.GetGameMapWidth() || y >= this.scene.sceneHeight {
		return false
	}
	t := this.scene.GetGameMapGridType(x, y)
	return t != usercmd.CellType_Wall && t != usercmd.CellType_Box
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
	glog.Infof("[%v收到伤害] 当前血量hp:%v", this.name, this.hp)
}

// 玩家造成伤害或击杀，增加得分
func (this *ScenePlayer) AddScore(x uint32) {
	atomic.StoreUint32(&this.score, this.score+x)
}

// ---------------------------------------------------------- //

// 发送场景同步信息
func (this *ScenePlayer) SendSceneMessage() {
	ret := &usercmd.RetUpdateSceneMsg{}
	// 场景内所有的玩家信息
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
	ret.Id = this.id
	ret.X = float32(this.curPos.X)
	ret.Y = float32(this.curPos.Y)

	this.self.SendCmd(usercmd.MsgTypeCmd_SceneSync, ret)
}

func (this *ScenePlayer) Update() {

}

// 玩家死亡处理
func (this *ScenePlayer) Death() {
	this.scene.DelPlayer(this)             // 场景中删除
	this.self.room.RemovePlayer(this.self) // 把玩家从房间中删除

}
