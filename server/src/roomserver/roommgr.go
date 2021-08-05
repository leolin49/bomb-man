package main

import "sync"

const (
	RoomType_1v1Room = 1
)

type RoomMgr struct {
	mutex    sync.Mutex
	roomList []*Room
	maxNum   int
	curNum   uint32 // 当前最后一个不满人的房间id
	endchan  chan uint32
}

var roommgr *RoomMgr

func RoomMgr_GetMe() *RoomMgr {
	if roommgr == nil {
		roommgr = &RoomMgr{
			roomList: make([]*Room, 10),
			endchan:  make(chan uint32, 100),
			maxNum:   10,
			curNum:   1,
		}

	}
	return roommgr
}

// logicserver通过rpc获取房间id时调用
// func (this *RoomMgr) ApplyRoomId() {
// 	this.mutex.Lock()
// 	defer this.mutex.Unlock()
// 	var room *Room
// 	if this.roomList[this.curNum] == nil {
// 		room = NewRoom(RoomType_1v1Room, this.curNum)
// 		//
// 		room.curPlayerNum++
// 		room.id = this.curNum
// 		//
// 		this.roomList[this.curNum] = room
// 	} else { // 当前存在不满人的房间
// 		room = this.roomList[this.curNum]
// 		room.curPlayerNum++
// 		defer func() {
// 			if room.IsFull() { // 分配后房间满员
// 				this.curNum++
// 			}
// 		}()
// 	}
// }
