package main

import (
	"paopao/server/src/common"
	"sync/atomic"
)

type Scene struct {
	players     map[uint64]*ScenePlayer
	room        *Room
	Obstacle    *map[uint32]*common.Obstacle
	Box         *map[uint32]*common.Box
	BombMap     *map[uint32]*Bomb
	sceneWidth  float64
	sceneHeight float64

	bombNum uint32 // 炸弹编号
}

// 场景信息初始化
func (this *Scene) Init(room *Room) {
	// 房间指针
	this.room = room
	// 全部玩家列表
	this.players = make(map[uint64]*ScenePlayer)
	//
	this.bombNum = 0
}

func (this *Scene) Update() {
	// TODO
}

// 场景内添加一个玩家
func (this *Scene) AddPlayer(player *PlayerTask) {
	if player != nil {
		this.players[player.id] = NewScenePlayer(player, this)
	}
}

// 添加一个炸弹
func (this *Scene) AddBomb(bomb *Bomb) {
	// TODO 场景添加炸弹
}

// 删除一个炸弹（炸弹爆炸）
func (this *Scene) DelBomb() {
	// TODO 场景删除炸弹
}

// 获取下一个炸弹的编号
func (this *Scene) GetNextBombId() uint32 {
	return atomic.AddUint32(&this.bombNum, 1)
}

// 保证位置不超出地图范围
func (this *Scene) BorderCheck(pos *common.Position) {
	if pos.X < 0 {
		pos.X = 0
	} else if pos.X >= this.sceneWidth {
		pos.X = this.sceneWidth - 0.01
	}
	if pos.Y < 0 {
		pos.Y = 0
	} else if pos.Y >= this.sceneHeight {
		pos.Y = this.sceneHeight - 0.01
	}
}

// 定时发送场景信息，包括各个玩家的信息
func (this *Scene) SendRoomMessage() {
	for _, player := range this.players {
		player.SendSceneMessage()
	}
}
