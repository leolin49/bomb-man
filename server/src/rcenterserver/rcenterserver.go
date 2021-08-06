package main

import (
	"flag"
	"fmt"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

// 在rcenterserver维护各个roomserver基本数据信息，目的主要是【生成roomid】和【负载均衡】
type RoomServerInfo struct {
	PlayerNum uint32
	RoomNum   uint32
	Load      uint32
	CurRoomId uint32
}

type RCenterServer struct {
	gonet.Service
	// rpcser  *gonet.TcpServer
	// sockser *gonet.TcpServer
	// key:address	val:RoomServerInfo
	roomMap map[string]*RoomServerInfo
}

var server *RCenterServer

func RCenterServer_GetMe() *RCenterServer {
	if server == nil {
		server = &RCenterServer{
			// rpcser:  &gonet.TcpServer{},
			// sockser: &gonet.TcpServer{},
			roomMap: make(map[string]*RoomServerInfo),
		}
		server.Derived = server
	}
	return server
}

func (this *RCenterServer) Init() bool {
	if !RpcServerStart() {
		glog.Errorln("[RCenterServer] rpc service error")
		return false
	}
	if !GrpcServerStart() {
		glog.Errorln("[RCenterServer] grpc service error")
		return false
	}
	return true
}

func (this *RCenterServer) Final() bool {
	return true
}

func (this *RCenterServer) Reload() {

}

func (this *RCenterServer) MainLoop() {
	time.Sleep(time.Second)
}

func (this *RCenterServer) AddRoomServer(ip string, port uint32) {
	key := fmt.Sprintf("%v:%v", ip, port)
	this.roomMap[key] = &RoomServerInfo{}
}

func (this *RCenterServer) UpdateRoomServer(info usercmd.RoomServerInfo) {
	key := fmt.Sprintf("%v:%v", info.Ip, info.Port)
	this.roomMap[key].RoomNum = info.RoomNum
	this.roomMap[key].PlayerNum = info.PlayerNum
	this.roomMap[key].CurRoomId = info.CurRoomId
	// TODO 计算负载（负载计算方法）
}

var config = flag.String("config", "", "config path")

func main() {
	flag.Parse()
	env.Load(*config)
	defer glog.Flush()
	RCenterServer_GetMe().Main()
	glog.Info("[Close] RCenterServer closed.")
}
