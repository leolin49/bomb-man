package main

import (
	"flag"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"
	"time"

	"github.com/golang/glog"
)

type RCenterServer struct {
	gonet.Service
	// rpcser  *gonet.TcpServer
	// sockser *gonet.TcpServer
	// key:address	val:RoomServerInfo
}

var server *RCenterServer

func RCenterServer_GetMe() *RCenterServer {
	if server == nil {
		server = &RCenterServer{
			// rpcser:  &gonet.TcpServer{},
			// sockser: &gonet.TcpServer{},
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

var config = flag.String("config", "", "config path")

func main() {
	flag.Parse()
	env.Load(*config)
	defer glog.Flush()
	RCenterServer_GetMe().Main()
	glog.Info("[Close] RCenterServer closed.")
}
