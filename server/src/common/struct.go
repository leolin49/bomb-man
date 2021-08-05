package common

type RoomTokenInfo struct {
	UserId   uint32
	UserName string
	RoomId   uint32
}

type Position struct {
	X float64
	Y float64
}

type Obstacle struct {
	Id  uint32
	Pos Position
}
