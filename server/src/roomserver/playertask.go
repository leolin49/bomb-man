package main

import (
	"encoding/json"
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

	score uint32 // 得分
}

func NewPlayerTask(conn net.Conn) *PlayerTask {
	m := &PlayerTask{
		tcptask:    *gonet.NewTcpTask(conn),
		activeTime: time.Now(),
	}
	m.tcptask.Derived = m
	return m
}

func (this *PlayerTask) SetUserInfo(info common.RoomTokenInfo) {
	this.id = info.UserId
	this.name = info.UserName
	this.score = 0
}

func (this *PlayerTask) Start() {
	this.tcptask.Start()
}

func (this *PlayerTask) OnClose() {
	this.tcptask.Close()
	if this.room != nil {
		this.room.RemovePlayer(this)
		this.room = nil
	}
	PlayerTaskManager_GetMe().Remove(this)
}

func (this *PlayerTask) ParseMsg(data []byte, flag byte) bool {

	this.activeTime = time.Now()

	info := usercmd.CmdHeader{}
	err := json.Unmarshal(data, &info)
	if err != nil {
		glog.Errorln("[json解析失败] ", err)
	}
	cmd := info.Cmd

	// 验证登录
	if !this.tcptask.IsVerified() {
		if cmd != usercmd.MsgTypeCmd_Login {
			glog.Errorf("[RoomServer Login] not a login instruction ", this.RemoteAddr())
			return false
		}
		// revCmd, ok := common.DecodeCmd(data, flag, &usercmd.UserLoginInfo{}).(*usercmd.UserLoginInfo)
		revCmd := &usercmd.UserLoginInfo{}
		err := json.Unmarshal([]byte(info.Data), revCmd)
		if err != nil {
			glog.Errorln(err)
			// this.retErrorMsg(common.ErrorCodeRoom)
			return false
		}
		glog.Infoln("[RoomServer Login] recv a login request ", this.RemoteAddr())
		// 解析token
		glog.Infoln(revCmd.Token)
		info, err := common.ParseRoomToken(revCmd.Token)
		if err != nil {
			glog.Errorln("[MsgTypeCmd_Login] parse room token error:", err)
			return false
		}

		key := info.UserName + "_roomtoken"
		// glog.Infoln(key)
		// glog.Infoln(info.UserId)
		// glog.Infoln(info.RoomId)
		token := RedisManager_GetMe().Get(key)
		if len(token) == 0 { // token不存在或者token过期
			glog.Errorln("[MsgTypeCmd_Login] token错误或者已经过期")
			PlayerTaskManager_GetMe().Remove(this) // 将playertask从manager中移除
			this.retErrorMsg(common.ErrorCodeInvalidToken)
			this.OnClose() // 断开连接
			return false
		}

		// --------验证通过-------------
		this.tcptask.Verify()
		this.SetUserInfo(*info) // 初始化playertask中的玩家信息

		room := RoomManager_GetMe().GetRoomById(info.RoomId)
		if room == nil { // 当前玩家为房间的第一位玩家，创建房间
			room = RoomManager_GetMe().NewRoom(RoomType_1v1, info.RoomId)
		}
		err = room.AddPlayer(this)
		if err != nil {
			glog.Errorln("[Enter Room] need retry")
		}
		this.retErrorMsg(common.ErrorCodeSuccess) //////////////////////
		return true
	}

	// TODO 加载地图信息
	if cmd == usercmd.MsgTypeCmd_NewScene {
		this.room.scene.gameMap = &GameMap{}

		// if common.DecodeGoCmd(data, flag, this.room.scene.gameMap) != nil {
		// 	glog.Infof("[解析游戏地图失败] %v load game map failed", this.id)
		// 	return false
		// }
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
		json.Unmarshal([]byte(info.Data), revCmd)
		// if common.DecodeGoCmd(data, flag, revCmd) != nil {
		// 	return false
		// }
		glog.Infof("[%v收到请求移动的指令] revCmd.Way=%v", this.name, revCmd.Way)
		if this.room == nil || this.room.IsClosed() {
			glog.Infoln("[收到请求移动的指令] 房间不存在")
			return false
		}
		if !this.room.IsStart() { // 游戏未开始
			glog.Infoln("[收到请求移动的指令] 游戏未开始")
			return false
		}
		this.room.chan_PlayerOp <- &PlayerOp{uid: this.id, op: PlayerMoveOp, msg: revCmd}

	case usercmd.MsgTypeCmd_PutBomb: // 放炸弹
		revCmd := &usercmd.MsgPutBomb{}
		json.Unmarshal([]byte(info.Data), revCmd)
		// if common.DecodeGoCmd(data, flag, revCmd) != nil {
		// 	return false
		// }
		glog.Infof("[%v收到请求放炸弹的指令]", this.name)
		if this.room == nil || this.room.IsClosed() {
			glog.Infoln("[收到请求放炸弹的指令] 房间不存在")
			return false
		}
		if !this.room.IsStart() { // 游戏未开始
			glog.Infoln("[收到请求放炸弹的指令] 游戏未开始")
			return false
		}
		this.room.chan_PlayerOp <- &PlayerOp{uid: this.id, op: PlayerPutBombOp, msg: revCmd}

	default:
		return false
	}

	return true
}

func (this *PlayerTask) SendCmd(cmd usercmd.MsgTypeCmd, msg common.Message) {
	data, ok := common.EncodeToBytesJson(uint16(cmd), msg)
	if !ok {
		glog.Errorf("[PlayerTask] send error cmd: %v, len: %v", cmd, len(data))
		return
	}
	// glog.Infoln(string(data))
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
