package main

import (
	"math"
	"math/rand"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"sync/atomic"
	"time"
)

// 初始数值
const (
	INIT_POWER = 1 // 初始炸弹威力
	INIT_SPEED = 3 // 初始移动速度

	BOMB_MAXTIME = 4 // 炸弹的持续时间

	HurtScore = 1 // 造成伤害得分
	KillScore = 3 // 击杀得分
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
	speed   uint32           // 移动速度

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
		curPos:  &common.Position{X: 0, Y: 0},
		hp:      3,
		curbomb: 0,
		maxbomb: 1,
		power:   INIT_POWER,
		speed:   INIT_SPEED,

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
	bomb := NewBomb(this, this.scene.GetNextBombId())
	this.scene.AddBomb(bomb)
	go bomb.CountDown()

	atomic.StoreUint32(&this.curbomb, 1)

	return true
}

// 移动
func (this *ScenePlayer) Move(msg *usercmd.MsgMove) {
	// 计算下一个位置
	this.CaculateNext(float64(msg.Way))
	// 边界检查
	this.BorderCheck(this.nextPos)
	// TODO 障碍物(墙体，...)检查
	// ...

	this.curPos = this.nextPos
}

func (this *ScenePlayer) CaculateNext(direct float64) {
	this.nextPos.X = this.curPos.X + math.Sin(direct*math.Pi/180)*float64(this.speed)
	this.nextPos.Y = this.curPos.Y + math.Cos(direct*math.Pi/180)*float64(this.speed)
}

func (this *ScenePlayer) BorderCheck(pos *common.Position) {
	if pos.X < 0 {
		pos.X = 0
	} else if w := this.scene.sceneWidth; pos.X >= w {
		pos.X = w - 0.1
	}
	if pos.Y < 0 {
		pos.Y = 0
	} else if h := this.scene.sceneHeight; pos.Y >= h {
		pos.Y = h - 0.1
	}
}

// 收到伤害
func (this *ScenePlayer) BeHurt(attacker *ScenePlayer) {
	if atomic.StoreInt32(&this.hp, -1); this.hp <= 0 {
		// TODO 玩家角色死亡
		info := &usercmd.RetRoleDeath{
			KillName: attacker.name,
			KillId:   attacker.id,
			LiveTime: 0,
			Score:    this.score,
		}
		this.self.SendCmd(usercmd.MsgTypeCmd_Death, info)
	}
}

// 玩家造成伤害或击杀，增加得分
func (this *ScenePlayer) AddScore(x uint32) {
	atomic.StoreUint32(&this.score, x)
}

// ---------------------------------------------------------- //

func (this *ScenePlayer) SendSceneMessage() {
	ret := &usercmd.RetUpdateSceneMsg{}
	// 场景内所有的玩家信息
	for _, player := range this.scene.players {
		ret.Players = append(ret.Players, &usercmd.ScenePlayer{
			Id:      player.id,
			BombNum: player.curbomb,
			Power:   player.power,
			Speed:   player.speed,
			State:   uint32(player.hp),
			X:       float32(player.curPos.X),
			Y:       float32(player.curPos.Y),
			IsMove:  player.isMove,
		})
	}
	// 场景内所有的炸弹信息
	for _, bomb := range *this.scene.BombMap {
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
