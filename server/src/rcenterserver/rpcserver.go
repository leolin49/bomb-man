package main

import (
	"net"
	"net/rpc"
	"paopao/server-base/src/base/env"
	"paopao/server/src/common"
	"paopao/server/usercmd"

	"github.com/golang/glog"
)

// rpc
const RpcServiceName = "RPC.GetRoomServerInfo"

type RpcRoomService struct {
}

func (r *RpcRoomService) RetRoom(request *usercmd.ReqIntoRoom, reply *usercmd.RetIntoFRoom) error {
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

func RpcServerStart() bool {
	err := rpc.RegisterName(RpcServiceName, new(RpcRoomService))
	if err != nil {
		glog.Errorln("[RpcServerStart] rpc RegisterName error:", err)
		return false
	}
	glog.Infoln("[RpcServerStart] address: ", env.Get("rcenter", "rpc_server"))
	listener, err := net.Listen("tcp", env.Get("rcenter", "rpc_server"))
	if err != nil {
		glog.Errorln("[RCenterServer] listen error:", err)
		return false
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				glog.Errorln("[RCenterServer] accept error:", err)
				continue
			}
			go rpc.ServeConn(conn)
		}
	}()
	glog.Infoln("[RpcServerStart] rpc service start success")
	return true
}
