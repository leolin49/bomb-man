package main

import (
	"sync"
	"sync/atomic"

	"github.com/golang/glog"
)

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
		go roommgr.Start()
	}
	return roommgr
}

func (this *RoomManager) Start() {
	for {
		select {
		case rid := <-this.endchan:
			this.mutex.Lock()
			delete(this.roomMap, rid)
			glog.Infof("[Game] room %v end", rid)
			this.mutex.Unlock()
		}
	}
}

func (this *RoomManager) UpdateNextRoomId() {
	atomic.StoreUint32(&this.curNum, this.curNum+1)
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
