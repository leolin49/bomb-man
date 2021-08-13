package main

import (
	"flag"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"

	"github.com/golang/glog"
)

type RoomServer struct {
	gonet.Service
	tcpser  *gonet.TcpServer
	version uint32
}

var server *RoomServer

func RoomServer_GetMe() *RoomServer {
	if server == nil {
		server = &RoomServer{
			tcpser: &gonet.TcpServer{},
		}
		server.Derived = server
	}
	return server
}

func (this *RoomServer) Init() bool {
	if !RoomGrpcClient_GetMe().Init() {
		glog.Errorln("[RoomServer] room grpc client init error")
		return false
	}
	go func() {
		err := this.tcpser.Bind(":" + *port)
		if err != nil {
			glog.Errorln("[RoomServer] Binding port failed")
			return
		}
	}()
	glog.Infof("[RoomServer] roomserver init success")
	return true
}

func (this *RoomServer) Final() bool {
	this.tcpser.Close()
	RedisManager_GetMe().Stop()
	RoomGrpcClient_GetMe().RemoveRoomServer()
	return true
}

func (this *RoomServer) Reload() {
}

func (this *RoomServer) MainLoop() {
	conn, err := this.tcpser.Accept()
	if err != nil {
		return
	}
	playerTask := NewPlayerTask(conn)
	PlayerTaskManager_GetMe().Add(playerTask)
	glog.Infof("[RoomServer] NewPlayerTask %v success", conn.RemoteAddr().String())
	playerTask.Start()
}

var (
	port   = flag.String("port", "13000", "roomserver listen port")
	config = flag.String("config", "", "config json file path")
)

func main() {
	flag.Parse()
	env.Load(*config)
	glog.Infoln("[debug] port:", *port)
	defer glog.Flush()
	RoomServer_GetMe().Main()
}
