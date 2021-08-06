package main

import "sync"

type RoomManager struct {
	mutex   sync.Mutex
	roomMap map[uint32]*Room
	maxNum  int
	curNum  uint32 // 当前最后一个不满人的房间id
	endchan chan uint32
}

var roommgr *RoomManager

func RoomManager_GetMe() *RoomManager {
	if roommgr == nil {
		roommgr = &RoomManager{
			roomMap: make(map[uint32]*Room),
			endchan: make(chan uint32, 100),
			maxNum:  10,
			curNum:  1,
		}

	}
	return roommgr
}

func (this *RoomManager) AddRoom(room *Room) (*Room, bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	oldroom, ok := this.roomMap[room.id]
	if ok {
		return oldroom, false
	}
	this.roomMap[this.curNum] = room
	return room, true
}

// 新增房间
func (this *RoomManager) NewRoom(rtype, rid uint32) *Room {
	room, ok := this.AddRoom(NewRoom(rtype, rid))
	if ok {
		//...
	}
	return room
}

func (this *RoomManager) GetRoomById(rid uint32) *Room {
	this.mutex.Lock()
	room, ok := this.roomMap[rid]
	this.mutex.Unlock()
	if !ok {
		return nil
	}
	return room
}

func (this *RoomManager) GetRoomList() (rooms []*Room) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, room := range this.roomMap {
		rooms = append(rooms, room)
	}
	return rooms
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
