package main

import (
	"errors"
	"paopao/server/usercmd"
	"sync"
	"time"

	"github.com/golang/glog"
)

// 房间类型
const (
	ROOMTYPE_1V1 = 1
)

type Room struct {
	scene *Scene // 场景信息

	mutex        sync.Mutex
	id           uint32                 //房间id
	roomType     uint32                 //房间类型
	players      map[uint64]*PlayerTask //房间内的玩家
	curPlayerNum uint32                 //当前房间内玩家数
	bombCount    uint32
	isStart      bool
	stopCh       chan bool
	isStop       bool
	iscustom     bool
	timeloop     uint64

	maxPlayerNum uint32 // max player number. default :8
	startTime    uint64 // 开始时间
	totalTime    uint64 // in second
	endTime      uint64 // 结束时间
	endchan      chan bool

	chan_PlayerOp chan *PlayerOp
}

func NewRoom(roomtype, roomid uint32) *Room {
	room := &Room{
		id:            roomid,
		roomType:      roomtype,
		players:       make(map[uint64]*PlayerTask),
		curPlayerNum:  0,
		maxPlayerNum:  2,
		isStart:       false,
		isStop:        false,
		startTime:     uint64(time.Now().Unix()),
		endchan:       make(chan bool),
		chan_PlayerOp: make(chan *PlayerOp, 500),
	}
	room.scene = NewScene(room) // 初始化场景信息
	glog.Infof("[NewRoom] roomtype:%v, roomid:%v", roomtype, roomid)
	return room
}

// 玩家进入房间
func (this *Room) AddPlayer(player *PlayerTask) error {
	this.mutex.Lock()
	if this.curPlayerNum >= this.maxPlayerNum {
		return errors.New("room is full")
	}
	// 更新房间信息
	this.curPlayerNum++
	player.room = this
	this.players[player.id] = player
	glog.Infof("[房间] 玩家进入房间  username:%v, uid:%v", player.name, player.id)
	this.scene.AddPlayer(player) // 将玩家添加到场景
	// 房间内玩家数量达到最大，自动开始游戏
	if this.curPlayerNum == this.maxPlayerNum {
		glog.Infoln("[房间] 玩家数量：", len(this.players))
		glog.Infoln("[游戏开始] 玩家列表：")
		for _, v := range this.players {
			glog.Infof("username:%v, uid:%v", v.scenePlayer.name, v.scenePlayer.id)
		}
		RoomManager_GetMe().UpdateNextRoomId() // 房间id++
		go this.StartGame()
	}
	this.mutex.Unlock()

	return nil
}

// 将玩家移除出房间
func (this *Room) RemovePlayer(player *PlayerTask) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	delete(this.players, player.id)
	this.curPlayerNum--
	return nil
}

func (this *Room) StartGame() {
	this.isStart = true
	this.GameLoop()
}

func (this *Room) Update() {
	this.scene.Update()
}

func (this *Room) IsFull() bool {
	return this.curPlayerNum == this.maxPlayerNum
}

func (this *Room) IsClosed() bool {
	// return atomic.LoadInt32(&this.isStop) != 0
	return this.isStop
}

// 房间结束
func (this *Room) Close() {
	if !this.isStop {
		// TODO房间结束处理
		this.isStop = true
		RoomManager_GetMe().endchan <- this.id
	}
}

func (this *Room) GameLoop() {

	// TODO 加载地图信息
	this.scene.RandGameMapData_AllSpace()

	timeTicker := time.NewTicker(time.Millisecond * 20) // 20ms
	stop := false
	for !stop {
		select {
		// 定时同步
		case <-timeTicker.C:
			// 0.04s
			if this.timeloop%2 == 0 {
				this.Update()
			}
			// 0.1s
			if this.timeloop%20 == 0 {
				this.scene.SendRoomMessage()
			}
			// TODO 游戏达到最长时间，自动结束
			this.timeloop++
			if this.isStop {
				stop = true
			}
		// 玩家主动操作
		case playerop := <-this.chan_PlayerOp:
			switch playerop.op {
			// 移动操作
			case PlayerMoveOp:
				glog.Errorf("[%v] execute move cmd", playerop.uid)
				req, ok := playerop.msg.(*usercmd.MsgMove)
				if !ok {
					glog.Errorln("[Move] move arg error")
					return
				}
				this.scene.players[playerop.uid].Move(req)
				break
			// 放置炸弹
			case PlayerPutBombOp:
				glog.Errorf("[%v] execute put bomb cmd", playerop.uid)
				req, ok := playerop.msg.(*usercmd.MsgPutBomb)
				if !ok {
					glog.Errorln("[PutBomb] put bomb arg error")
					return
				}
				this.scene.players[playerop.uid].PutBomb(req)
				break
			}
		}
	}
	this.Close()
}

func (this *Room) PlayerSceneSync(task *PlayerTask, opts *PlayerOp) {

}
