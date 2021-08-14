package main

type GridType uint32

const (
	GridType_Space    = GridType(iota) // 空地
	GridType_Obstacle                  // 不可摧毁的障碍物
	GridType_Box                       // 箱子（可摧毁）
	GridType_Bomb                      // 炸弹
)

type GameMap struct {
	Width    uint32
	Height   uint32
	MapArray [][]GridType
}

func (this *GameMap) GetGridByPos(x, y uint32) GridType {
	return this.MapArray[x][y]
}

// func (this *GameMap) CanPass(x, y int) bool {
// 	return this.MapArray[x][y] != GridType_Box &&
// 		this.MapArray[x][y] != GridType_Obstacle
// }

// func (this *GameMap) GetWidth() uint32 {
// 	return this.Width
// }

// func (this *GameMap) GetHeight() uint32 {
// 	return this.Height
// }
