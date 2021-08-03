package main

import (
	"flag"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"

	"github.com/golang/glog"
)

type RoomServer struct {
	gonet.Service
	roomser *gonet.TcpServer
	version uint32
}

var server *RoomServer

func RoomServer_GetMe() *RoomServer {
	if server == nil {
		server = &RoomServer{
			roomser: &gonet.TcpServer{},
		}
		server.Derived = server
	}
	return server
}

func (this *RoomServer) Init() bool {

	err := this.roomser.Bind(env.Get("room", "tcp_server"))
	if err != nil {
		glog.Errorln("[RoomServer] Binding port failed")
		return false
	}
	return true
}

func (this *RoomServer) Final() bool {
	return true
}

func (this *RoomServer) Reload() {
}

func (this *RoomServer) MainLoop() {
}

var (
	port   = flag.String("port", "8000", "logicserver listen port")
	config = flag.String("config", "", "config json file path")
)

func main() {
	// TODO
}
