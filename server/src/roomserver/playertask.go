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
	gonet.TcpTask
	// udptask *snet.Session
	isUdp       bool
	key         string
	id          uint64
	name        string
	room        *Room
	scenePlayer *ScenePlayer
	uobjs       []uint32
	activeTime  time.Time
	onlineTime  int64
}

func NewPlayerTask(conn net.Conn) *PlayerTask {
	m := &PlayerTask{
		TcpTask:    *gonet.NewTcpTask(conn),
		activeTime: time.Now(),
	}
	m.Derived = m
	return m
}

func (this *PlayerTask) OnClose() {
	this.room = nil
}

func (this *PlayerTask) ParseMsg(data []byte, flag byte) bool {

	this.activeTime = time.Now()

	cmd := usercmd.MsgTypeCmd(common.GetCmd(data))
	if this.IsVerified() {
		switch cmd {
		case usercmd.MsgTypeCmd_Login: // 玩家连接roomserver
			revCmd, ok := common.DecodeCmd(data, flag, &usercmd.UserLoginInfo{}).(*usercmd.UserLoginInfo)
			if !ok {
				return false
			}
			// 解析token
			info, err := common.ParseRoomToken(*revCmd.Token)
			if err != nil {
				glog.Errorln("[MsgTypeCmd_Login] parse room token error:", err)
				return false
			}
			key := info.UserName + "_roomtoken"
			token := common.RedisMgr.Get(key)
			if len(token) == 0 { // token不存在或者token过期
				glog.Errorln("[MsgTypeCmd_Login] token expired")
				return false
			}
			this.Verify() // 验证通过
			// roommgr.roomList[info.RoomId].AddPlayer()

		case usercmd.MsgTypeCmd_Move: // 移动
			// TODO
		case usercmd.MsgTypeCmd_PutBomb: // 放炸弹
			// TODO
		case usercmd.MsgTypeCmd_Death: // 死亡
			// TODO
		case usercmd.MsgTypeCmd_HeartBeat: // 心跳
			// TODO
		default:
		}
	}

	return true
}
