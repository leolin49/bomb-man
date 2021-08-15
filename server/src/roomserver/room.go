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
	RoomType_1v1 = 1
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

	maxPlayerNum uint32    // max player number. default :8
	startTime    time.Time // 开始时间
	totalTime    uint64    // in second
	endTime      time.Time // 结束时间
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
	defer this.mutex.Unlock()
	if this.curPlayerNum >= this.maxPlayerNum {
		return errors.New("room is full")
	}
	// 更新房间信息
	this.curPlayerNum++
	player.room = this
	this.players[player.id] = player
	glog.Infof("[房间] 玩家[%v]进入[%v]房间  ", player.name, this.id)
	this.scene.AddPlayer(player) // 将玩家添加到场景

	// 房间内玩家数量达到最大，自动开始游戏
	if this.curPlayerNum == this.maxPlayerNum {
		glog.Infoln("[房间] 玩家数量：", len(this.players))
		glog.Infoln("[游戏开始] 玩家列表：")
		for _, v := range this.players {
			glog.Infof("username:%v, uid:%v", v.scenePlayer.name, v.scenePlayer.id)
		}
		RoomManager_GetMe().UpdateNextRoomId() // 房间id++

		// 将当前房间内的所有玩家信息发送到客户端
		for _, pt := range this.players {
			info := &usercmd.RetUpdateSceneMsg{}
			info.Id = pt.id
			for _, sp := range this.scene.players {
				info.Players = append(info.Players, &usercmd.ScenePlayer{
					Id:      sp.id,
					BombNum: sp.curbomb,
					Power:   sp.power,
					Speed:   float32(sp.speed),
					State:   uint32(sp.hp),
					X:       float32(sp.curPos.X),
					Y:       float32(sp.curPos.Y),
					Score:   10,
					IsMove:  sp.isMove,
				})
			}
			pt.SendCmd(usercmd.MsgTypeCmd_SceneSync, info)
		}

		go this.StartGame()
	}

	return nil
}

// 将玩家移除出房间
func (this *Room) RemovePlayer(player *PlayerTask) error {
	if this == nil {
		return nil
	}
	this.mutex.Lock()
	glog.Warningln("[debug]Room.RemovePlayer() func")
	defer this.mutex.Unlock()
	delete(this.players, player.id)
	glog.Warningln("[debug]Room.RemovePlayer() func")
	return nil
}

func (this *Room) StartGame() {
	this.isStart = true
	this.startTime = time.Now()
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

func (this *Room) IsStart() bool {
	return this.isStart
}

// 房间结束
func (this *Room) Close() {
	if !this.isStop {
		// 房间结束处理
		this.isStop = true
		// 删除房间内玩家，断开所有连接
		for _, player := range this.players {
			player.OnClose()
		}
		glog.Infof("[房间%v游戏结束] 游戏持续时间:%v", this.id, this.totalTime)
		RoomManager_GetMe().endchan <- this.id
	}
}

func (this *Room) GameLoop() {

	// TODO 加载地图信息
	if !this.scene.gameMap.CustomizeInit() {
		glog.Errorln("[地图加载失败]")
		return
	}

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
			if this.timeloop%5 == 0 {
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
				// glog.Errorf("[%v] execute move cmd", playerop.uid)
				req, ok := playerop.msg.(*usercmd.MsgMove)
				if !ok {
					glog.Errorln("[Move] move arg error")
					return
				}
				this.scene.players[playerop.uid].Move(req)
			// 放置炸弹
			case PlayerPutBombOp:
				// glog.Errorf("[%v] execute put bomb cmd", playerop.uid)
				req, ok := playerop.msg.(*usercmd.MsgPutBomb)
				if !ok {
					glog.Errorln("[PutBomb] put bomb arg error")
					return
				}
				this.scene.players[playerop.uid].PutBomb(req)
			}
		case <-this.endchan:
			this.Close()
		}
	}
	this.Close()
}

func (this *Room) PlayerSceneSync(task *PlayerTask, opts *PlayerOp) {

}
