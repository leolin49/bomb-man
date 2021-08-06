package common

const (
	BoxGoodsBomb = 1 // 加炸弹上限
	BoxGoodsLife = 2 // 加生命
)

type RoomTokenInfo struct {
	UserId   uint32
	UserName string
	RoomId   uint32
}

type Position struct {
	X float64
	Y float64
}

// 障碍物
type Obstacle struct {
	Id  uint32
	Pos Position
}

// 宝箱
type Box struct {
	Goods uint32 // 物体类型
	Id    uint32
	Pos   Position // 位置
}
