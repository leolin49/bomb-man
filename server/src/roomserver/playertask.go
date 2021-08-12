package main

import (
	"net"
	"paopao/server-base/src/base/gonet"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

type PlayerOpType int

const (
	PlayerNoneOp = PlayerOpType(iota)
	PlayerMoveOp
	PlayerPutBombOp
)

type PlayerOp struct {
	uid uint64         // 玩家id
	op  PlayerOpType   // 操作类型
	msg common.Message // 其他信息
}

type PlayerTask struct {
	// udptask *snet.Session
	tcptask     gonet.TcpTask
	isUdp       bool
	key         string
	id          uint64
	name        string
	room        *Room
	scenePlayer *ScenePlayer
	uobjs       []uint32

	activeTime time.Time
	onlineTime int64

	moveWay int32 // 移动方向
	hasMove int32 // 是否移动
}

func NewPlayerTask(conn net.Conn) *PlayerTask {
	m := &PlayerTask{
		tcptask:    *gonet.NewTcpTask(conn),
		activeTime: time.Now(),
	}
	m.tcptask.Derived = m
	return m
}

func (this *PlayerTask) Start() {
	this.tcptask.Start()
}

func (this *PlayerTask) OnClose() {
	this.tcptask.Close()
	PlayerTaskManager_GetMe().Remove(this)
	this.room = nil
}

func (this *PlayerTask) ParseMsg(data []byte, flag byte) bool {

	this.activeTime = time.Now()

	cmd := usercmd.MsgTypeCmd(common.GetCmd(data))

	// 验证登录
	if !this.tcptask.IsVerified() {
		if cmd != usercmd.MsgTypeCmd_Login {
			glog.Errorf("[RoomServer Login] not a login instruction ", this.RemoteAddr())
			return false
		}
		revCmd, ok := common.DecodeCmd(data, flag, &usercmd.UserLoginInfo{}).(*usercmd.UserLoginInfo)
		if !ok {
			this.retErrorMsg(common.ErrorCodeRoom)
			return false
		}
		glog.Infoln("[RoomServer Login] recv a login request ", this.RemoteAddr())
		// 解析token
		info, err := common.ParseRoomToken(revCmd.Token)
		if err != nil {
			glog.Errorln("[MsgTypeCmd_Login] parse room token error:", err)
			return false
		}
		key := info.UserName + "_roomtoken"
		token := common.RedisMgr.Get(key)
		if len(token) == 0 { // token不存在或者token过期
			glog.Errorln("[MsgTypeCmd_Login] token expired")
			this.retErrorMsg(common.ErrorCodeInvalidToken)
			return false
		}
		this.tcptask.Verify() // 验证通过
		room := RoomManager_GetMe().GetRoomById(info.RoomId)
		if room == nil { // 当前玩家为房间的第一位玩家，创建房间
			room = RoomManager_GetMe().NewRoom(ROOMTYPE_1V1, info.RoomId)
		}
		err = room.AddPlayer(this)
		if err != nil {
			glog.Errorln("[Enter Room] need retry")
		}
		this.retErrorMsg(common.ErrorCodeSuccess) //////////////////////
		return true
	}

	// 加载地图信息
	if cmd == usercmd.MsgTypeCmd_NewScene {
		this.room.scene.gameMap = &usercmd.MapVector{}
		if common.DecodeGoCmd(data, flag, this.room.scene.gameMap) != nil {
			return false
		}
		glog.Infof("[MsgTypeCmd_NewScene] %v load game map success", this.id)
		return true
	}

	// TODO 心跳
	if cmd == usercmd.MsgTypeCmd_HeartBeat {
		this.AsyncSend(data, flag)
		return true
	}

	switch cmd {
	case usercmd.MsgTypeCmd_Move: // 移动
		revCmd := &usercmd.MsgMove{}
		if common.DecodeGoCmd(data, flag, revCmd) != nil {
			return false
		}
		glog.Infoln("[收到请求移动的指令] revCmd.Way = ", revCmd.Way)
		if this.room == nil || this.room.IsClosed() {
			glog.Infoln("[收到请求移动的指令] 房间不存在")
			return false
		}
		this.room.chan_PlayerOp <- &PlayerOp{uid: this.id, op: PlayerMoveOp, msg: revCmd}
	case usercmd.MsgTypeCmd_PutBomb: // 放炸弹
		revCmd := &usercmd.MsgPutBomb{}
		if common.DecodeGoCmd(data, flag, revCmd) != nil {
			return false
		}
		glog.Infoln("[收到请求放炸弹的指令]")
		if this.room == nil || this.room.IsClosed() {
			glog.Infoln("[收到请求放炸弹的指令] 房间不存在")
			return false
		}
		this.room.chan_PlayerOp <- &PlayerOp{uid: this.id, op: PlayerPutBombOp, msg: revCmd}

	default:
		return false
	}

	return true
}

func (this *PlayerTask) SendCmd(cmd usercmd.MsgTypeCmd, msg common.Message) {
	data, ok := common.EncodeToBytes(uint16(cmd), msg)
	if !ok {
		glog.Errorf("[PlayerTask] send error cmd: %v, len: %v", cmd, len(data))
		return
	}
	this.AsyncSend(data, 0)
}

func (this *PlayerTask) AsyncSend(buffer []byte, flag byte) bool {
	if flag == 0 && len(buffer) > 1024 {
		// TODO优化：数据量大时，压缩后在发送
	}
	return this.tcptask.AsyncSend(buffer, flag)
}

func (this *PlayerTask) retErrorMsg(ecode uint32) {
	retCmd := &usercmd.RetErrorMsgCmd{
		RetCode: ecode,
	}
	this.SendCmd(usercmd.MsgTypeCmd_ErrorMsg, retCmd)
}

func (this *PlayerTask) RemoteAddr() string {
	return this.tcptask.RemoteAddr()
}

func (this *PlayerTask) LocalAddr() string {
	return this.tcptask.LocalAddr()
}
