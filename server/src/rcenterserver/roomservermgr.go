package main

import (
	"fmt"
	"paopao/server/usercmd"
)

// 在rcenterserver维护各个roomserver基本数据信息，目的主要是【生成roomid】和【负载均衡】
type RoomServerInfo struct {
	PlayerNum uint32
	RoomNum   uint32
	Load      uint32
	CurRoomId uint32
}

type RoomServerManager struct {
	roomMap map[string]*RoomServerInfo
}

var mRoomServerMgr *RoomServerManager

func RoomServerManager_GetMe() *RoomServerManager {
	if mRoomServerMgr == nil {
		mRoomServerMgr = &RoomServerManager{
			roomMap: make(map[string]*RoomServerInfo),
		}
	}
	return mRoomServerMgr
}

func (this *RoomServerManager) AddRoomServer(ip string, port uint32) {
	key := fmt.Sprintf("%v:%v", ip, port)
	this.roomMap[key] = &RoomServerInfo{}
}

func (this *RoomServerManager) UpdateRoomServer(info *usercmd.RoomServerInfo) {
	key := fmt.Sprintf("%v:%v", info.Ip, info.Port)
	this.roomMap[key].RoomNum = info.RoomNum
	this.roomMap[key].PlayerNum = info.PlayerNum
	this.roomMap[key].CurRoomId = info.CurRoomId
	// TODO 计算负载（负载计算方法）
}

func (this *RoomServerManager) DeleteRoomServer(ip string, port uint32) {
	key := fmt.Sprintf("%v:%v", ip, port)
	delete(this.roomMap, key)
}
