package common

const (
	BoxGoodsBomb = 1 // 加炸弹上限
	BoxGoodsLife = 2 // 加生命
)

const (
	PlayerTaskTimeOut = 20 // 玩家无操作超时时间（s）
)

type MoveWay byte

const (
	MoveWay_None = MoveWay(iota)
	MoveWay_Up
	MoveWay_Down
	MoveWay_Left
	MoveWay_Right
)

type RoomTokenInfo struct {
	UserId   uint32
	UserName string
	RoomId   uint32
}

// 坐标
type Position struct {
	X float64
	Y float64
}

// 格子
type GridPos struct {
	X uint32
	Y uint32
}

// 障碍物
type Obstacle struct {
	Id  uint32
	Pos GridPos
}

// 宝箱
type Box struct {
	Goods uint32 // 物体类型
	Id    uint32
	Pos   GridPos
}
