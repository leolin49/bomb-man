package main

import "paopao/server/src/common"

type Scene struct {
	players  map[uint64]*ScenePlayer
	room     *Room
	Obstacle *map[uint32]*common.Obstacle
	Box      *map[uint32]*common.Box
}

// 场景信息初始化
func (this *Scene) Init(room *Room) {
	this.players = make(map[uint64]*ScenePlayer)
	this.room = room
}

// 添加一个玩家
func (this *Scene) AddPlayer(player *PlayerTask) {
	this.players[player.id] = NewScenePlayer(player, this)
}

// 添加一个炸弹
func (this *Scene) AddBomb() {

}

// 删除一个炸弹（炸弹爆炸）
func (this *Scene) DelBomb() {

}

//
