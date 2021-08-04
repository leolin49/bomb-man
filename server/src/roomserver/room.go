package main

import "time"

const (
	RoomMaxNumber   = 10
	MaxPlayerInRoom = 2
)

type Room struct {
	//mutex    sync.RWMutex
	id           uint32                 //房间id
	roomType     uint32                 //房间类型
	players      map[uint32]*PlayerTask //房间内的玩家
	curPlayerNum uint32                 //当前房间内玩家数
	bombCount    uint32
	isStart      bool
	timeLoop     uint64
	stopCh       chan bool
	isStop       bool
	iscustom     bool

	max_num   uint32 //max player number. default :8
	startTime uint64
	totalTime uint64 //in second
	endchan   chan bool
}

func NewRoom(roomtype, roomid uint32) *Room {
	room := &Room{
		id:           roomid,
		roomType:     roomtype,
		curPlayerNum: 0,
		isStart:      false,
		isStop:       false,
		startTime:    uint64(time.Now().Unix()),
		endchan:      make(chan bool),
	}
	return room
}

func (this *Room) IsFull() bool {
	return this.curPlayerNum == MaxPlayerInRoom
}
