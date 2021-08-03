package main

import (
	"flag"
	"net"
	"net/rpc"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

// 在rcenterserver维护各个roomserver的信息，目的主要是负载均衡
type RoomServerLoad struct {
	ip   string
	port int
	load int
}

type RCenterServer struct {
	gonet.Service
	rpcser   *gonet.TcpServer
	sockser  *gonet.TcpServer
	roomList []RoomServerLoad
}

const RpcServiceName = "Rpc.RetInfoRoom"

var server *RCenterServer

func RCenterServer_GetMe() *RCenterServer {
	if server == nil {
		server = &RCenterServer{
			rpcser:  &gonet.TcpServer{},
			sockser: &gonet.TcpServer{},
		}
		server.Derived = server
	}
	return server
}

// rpc
type RetIntoRoom struct {
}

func (r *RetIntoRoom) RetRoom(request *usercmd.ReqIntoRoom, reply *usercmd.RetIntoFRoom) error {
	uid, username := request.GetUId(), request.GetUserName()

	glog.Infof("[Rpc get room]uid:%v, username:%v", uid, username)
	rpRoomId, rpAddr := uint32(0), "localhost:9494"
	reply.RoomId, reply.Addr = &rpRoomId, &rpAddr

	return nil
}

func Acc(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			glog.Errorln("[RCenterServer] accept error:", err)
			continue
		}
		go rpc.ServeConn(conn)
	}
}

func (this *RCenterServer) Init() bool {
	rpc.RegisterName(RpcServiceName, new(RetIntoRoom))
	listener, err := net.Listen("tcp", env.Get("rcenter", "server"))
	if err != nil {
		glog.Errorln("[RCenterServer] listen error:", err)
		return false
	}
	go Acc(listener)
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

func (this *RCenterServer) AddRoomServer()

var config = flag.String("config", "", "config path")

func main() {
	flag.Parse()
	env.Load(*config)
	defer glog.Flush()
	RCenterServer_GetMe().Main()

	glog.Info("[Close] RCenterServer closed.")
}
