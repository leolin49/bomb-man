package main

import "paopao/server/src/common"

type Scene struct {
	players  map[uint32]*ScenePlayer
	room     *Room
	Obstacle *map[uint32]*common.Obstacle
}

func (this *Scene) Init(room *Room) {
	this.room = room
}
