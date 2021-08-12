package main

import (
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"sync/atomic"

	"github.com/golang/glog"
)

type Scene struct {
	players     map[uint64]*ScenePlayer
	room        *Room
	ObstacleMap map[uint32]*common.Obstacle
	BoxMap      map[uint32]*common.Box
	BombMap     map[uint32]*Bomb
	sceneWidth  uint32
	sceneHeight uint32

	bombNum uint32 // 炸弹编号

	gameMap *usercmd.MapVector // 游戏地图信息
}

func NewScene(room *Room) *Scene {
	scene := &Scene{
		room:        room,
		players:     make(map[uint64]*ScenePlayer),
		ObstacleMap: make(map[uint32]*common.Obstacle),
		BoxMap:      make(map[uint32]*common.Box),
		bombNum:     0,
		gameMap:     nil,
	}
	return scene
}

// 场景信息初始化
func (this *Scene) Init(room *Room) {

}

// 加载地图数据
func (this *Scene) LoadGameMapData() bool {
	if this.gameMap == nil {
		glog.Errorln("[Scene] load game map error")
		return false
	}
	this.sceneHeight = uint32(len(this.gameMap.GetCol()[0].GetX()))
	this.sceneWidth = uint32(len(this.gameMap.GetCol()))
	// 纵坐标优先遍历
	var x, y uint32
	for y = 0; y < this.sceneWidth; y++ {
		for x = 0; x < this.sceneHeight; x++ {

			idx := x*this.sceneWidth + y // 二维转一维
			cellType := this.gameMap.GetCol()[y].GetX()[x]
			if cellType == usercmd.CellType_Wall {
				this.ObstacleMap[idx] = &common.Obstacle{
					Id: idx,
					Pos: common.GridPos{
						X: x,
						Y: y,
					},
				}
			} else if cellType == usercmd.CellType_Box {
				this.BoxMap[idx] = &common.Box{
					Id:    idx,
					Goods: 1, // TODO 宝箱里的物品
					Pos: common.GridPos{
						X: x,
						Y: y,
					},
				}
			}

		}
	}
	return true
}

// 自定义地图信息
func (this *Scene) RandGameMapData_AllSpace() {
	this.sceneHeight, this.sceneWidth = 10, 10
	var x, y uint32
	for y = 0; y < this.sceneWidth; y++ {
		for x = 0; x < this.sceneHeight; x++ {
			this.gameMap.GetCol()[y].GetX()[x] = usercmd.CellType_Space
		}
	}
}

func (this *Scene) Update() {
	// TODO
	for _, player := range this.players {
		player.Update()
	}
}

// 场景内添加一个玩家
func (this *Scene) AddPlayer(player *PlayerTask) {
	if player != nil {
		sp := NewScenePlayer(player, this)
		this.players[player.id] = sp
		player.scenePlayer = sp
	}
}

// 添加一个炸弹
func (this *Scene) AddBomb(bomb *Bomb) {
	this.BombMap[bomb.id] = bomb
	this.gameMap.Col[bomb.pos.Y].X[bomb.pos.X] = usercmd.CellType_Bomb
}

// 删除一个炸弹（炸弹爆炸）
func (this *Scene) DelBomb(bomb *Bomb) {
	delete(this.BombMap, bomb.id)
	this.gameMap.Col[bomb.pos.Y].X[bomb.pos.X] = usercmd.CellType_Space
	bomb = nil
}

// 获取下一个炸弹的编号
func (this *Scene) GetNextBombId() uint32 {
	return atomic.AddUint32(&this.bombNum, 1)
}

// 根据坐标返回地图上对应格子的当前类型（空地，墙体）
func (this *Scene) GetGameMapGridType(x, y uint32) usercmd.CellType {
	return this.gameMap.GetCol()[x].GetX()[y]
}

func (this *Scene) GetGameMapWidth() uint32 {
	return this.sceneWidth
}

// 定时发送场景信息，包括各个玩家的信息
func (this *Scene) SendRoomMessage() {
	for _, player := range this.players {
		player.SendSceneMessage()
	}
}
