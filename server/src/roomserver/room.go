package main

import (
	"errors"
	"sync"
	"time"

	"github.com/golang/glog"
)

const (
	RoomMaxNumber   = 10
	MaxPlayerInRoom = 2
)

const (
	ROOMTYPE_1V1 = 1
)

type Room struct {
	mutex        sync.Mutex
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
	glog.Infof("[NewRoom] roomtype:%v, roomid:%v", roomtype, roomid)
	return room
}

// 玩家进入房间
func (this *Room) AddPlayer(player *PlayerTask) error {
	this.mutex.Lock()
	if this.curPlayerNum >= MaxPlayerInRoom {
		return errors.New("room is full")
	}
	// 更新房间信息
	this.curPlayerNum++
	// 房间内玩家数量达到最大，自动开始游戏
	if this.curPlayerNum == MaxPlayerInRoom {
		RoomMgr_GetMe().curNum++ // 房间id++
		go this.StartGame()
	}
	this.mutex.Unlock()
	this.players[uint32(player.id)] = player
	// 更新玩家信息
	player.room = this

	return nil
}

func (this *Room) StartGame() {
	this.isStart = true
	this.GameLoop()
}

func (this *Room) GameLoop() {
	ticker := time.NewTicker(time.Millisecond * 10) // 10ms

	for !this.isStop {
		select {
		case <-ticker.C:
			// TODO 游戏状态同步
		}
	}
}

func (this *Room) IsFull() bool {
	return this.curPlayerNum == MaxPlayerInRoom
}

func (this *Room) Close() {
	this.isStop = true
}
