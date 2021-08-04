package main

import (
	"net/rpc"
	"paopao/server-base/src/base/env"
	"paopao/server/usercmd"

	"github.com/golang/glog"
)

type RcenterServiceClient struct {
	*rpc.Client
}

type RpcReqData struct {
	UserId   uint64
	UserName string
}

type RpcRspData struct {
	//...
	Address string `json:"roomserver_address"` // RoomServer 地址
	RoomId  uint32 `json:"room_id"`            // 房间 ID
	//...
}

const RpcServiceName = "RPC.GetRoomServerInfo"

func DialRcenterService(network, address string) (*RcenterServiceClient, error) {
	c, err := rpc.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return &RcenterServiceClient{Client: c}, nil
}

func (p *RcenterServiceClient) RequestService(request usercmd.ReqIntoRoom, reply *usercmd.RetIntoFRoom) error {
	return p.Client.Call(RpcServiceName, request, reply)
}

// 同步阻塞rpc
func RequestRpcService(request usercmd.ReqIntoRoom) *usercmd.RetIntoFRoom {
	rpcServer := env.Get("logic", "rcenter_rpc_server")
	client, err := DialRcenterService("tcp", rpcServer)
	defer client.Close()
	if err != nil {
		glog.Errorln("[LogicServer Rpc] Dial Failed, ", err)
		return nil
	}
	var respon usercmd.RetIntoFRoom
	err = client.RequestService(request, &respon)
	if err != nil {
		glog.Errorln("[LogicServer Rpc] request failed, ", err)
		return nil
	}
	return &respon
}

// 异步rpc请求
func AsynRequestRpcService(request usercmd.ReqIntoRoom) *usercmd.RetIntoFRoom {
	rpcServer := env.Get("logic", "rcenter_rpc_server")
	client, err := DialRcenterService("tcp", rpcServer)
	defer client.Close()
	if err != nil {
		glog.Errorln("[LogicServer Rpc] Dial Failed")
		return nil
	}
	call := client.Go(RpcServiceName, request, new(RpcRspData), nil)

	call = <-call.Done
	if err = call.Error; err != nil {
		glog.Errorln("[LogicServer Rpc] Asyn Request Failed")
		return nil
	}
	respon := call.Reply.(usercmd.RetIntoFRoom)
	return &respon
}
