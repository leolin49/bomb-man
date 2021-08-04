package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"paopao/server-base/src/base/env"
	"paopao/server-base/src/base/gonet"
	"paopao/server/src/common"
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
	rpcser  *gonet.TcpServer
	sockser *gonet.TcpServer
	// key:address	val:RoomServerInfo
	roomMap map[string]*RoomServerInfo
}

const RpcServiceName = "Rpc.RetInfoRoom"

var server *RCenterServer

func RCenterServer_GetMe() *RCenterServer {
	if server == nil {
		server = &RCenterServer{
			rpcser:  &gonet.TcpServer{},
			sockser: &gonet.TcpServer{},
			roomMap: make(map[string]*RoomServerInfo),
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
	rpError := uint32(0)
	var rpAddr string
	// TODO根据负载选择roomserver
	for k, v := range server.roomMap {
		_ = v
		rpAddr = k
		break
	}
	// 匹配成功后，rcenterserver会选择（如何选择？）一个roomserver并生成一个room id和一个token，
	// 而后将对应roomserver的地址以及token返回给logicserver，最后返回给用户
	var info common.RoomTokenInfo
	info.UserId = uint32(uid) // uint64->uint32
	info.UserName = username
	info.RoomId = server.roomMap[rpAddr].CurRoomId ////////////////////////////
	token, err := common.CreateRoomToken(info)
	if err != nil {
		glog.Errorln("[Rpc get room] create token error")
		return err
	}

	reply.Err, reply.Addr, reply.Key = &rpError, &rpAddr, &token

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

func (this *RCenterServer) AddRoomServer(ip string, port int) {
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
