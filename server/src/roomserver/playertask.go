package main

import (
	"net"
	"paopao/server-base/src/base/gonet"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

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
}

func NewPlayerTask(conn net.Conn) *PlayerTask {
	m := &PlayerTask{
		tcptask:    *gonet.NewTcpTask(conn),
		activeTime: time.Now(),
	}
	m.tcptask.Derived = m
	return m
}

func (this *PlayerTask) OnClose() {
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

	// 心跳
	if cmd == usercmd.MsgTypeCmd_HeartBeat {
		// TODO
	}

	switch cmd {
	case usercmd.MsgTypeCmd_Move: // 移动
		// TODO
		revCmd := &usercmd.MsgMove{}
		common.DecodeGoCmd(data, flag, revCmd)

	case usercmd.MsgTypeCmd_PutBomb: // 放炸弹
		// TODO
	case usercmd.MsgTypeCmd_Death: // 死亡
		// TODO
	default:
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
