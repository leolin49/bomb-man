package main

const (
	AoiGridSize = 2 // 每个aoi格子长度等于多少个游戏格子
)

type AoiGrid struct {
	index   uint32
	players map[uint64]*ScenePlayer
}

type AoiMap struct {
	mp [9]AoiGrid
}

// 根据玩家坐标计算其所在aoi格子的编号
func CaculateAoiGrid(x, y uint32) uint32 {
	return (y/AoiGridSize)*(MaxGridNumberX/AoiGridSize) + (x/AoiGridSize + 1)
}

func (this *AoiGrid) AddPlayer(player *ScenePlayer) {
	this.players[player.id] = player
}

func (this *AoiGrid) DelPlayer(player *ScenePlayer) {
	delete(this.players, player.id)
}

//

// // --------------------------------AOI（九宫格）-------------------------------- //
// // GetAoiGrid 获取玩家当前位置所处的AOI九宫格位置
// func (this *ScenePlayer) GetAoiGrid() (uint32, uint32) {
// 	x, y := this.GetCurrentGrid()
// 	x /= uint32(this.vision)
// 	y /= uint32(this.vision)
// 	return x, y
// }

// // GetAoiJgg 获取当前玩家所在整个九宫格中的所有玩家
// func (this *ScenePlayer) GetAoiJggAllPlayer(x, y uint32) []*ScenePlayer {
// 	var arr []common.GridPos
// 	arr = append(arr,
// 		common.GridPos{X: x - 1, Y: y - 1},
// 		common.GridPos{X: x, Y: y - 1},
// 		common.GridPos{X: x + 1, Y: y - 1},
// 		common.GridPos{X: x - 1, Y: y},
// 		common.GridPos{X: x, Y: y},
// 		common.GridPos{X: x + 1, Y: y},
// 		common.GridPos{X: x - 1, Y: y + 1},
// 		common.GridPos{X: x, Y: y + 1},
// 		common.GridPos{X: x + 1, Y: y + 1})
// 	res := make([]*ScenePlayer, 0)
// 	for i := 0; i < len(arr); i++ {

// 	}
// 	return res
// }
