package main

import (
	"context"
	"encoding/json"
	"paopao/server-base/src/base/env"
	"paopao/server/usercmd"
	"strconv"
	"time"

	"github.com/golang/glog"
	"google.golang.org/grpc"
)

type RoomGrpcClient struct {
	conn         *grpc.ClientConn
	mRouteClient usercmd.StreamRoomService_RouteClient
}

var mRoomGrpcClient *RoomGrpcClient

func RoomGrpcClient_GetMe() *RoomGrpcClient {
	if nil == mRoomGrpcClient {
		mRoomGrpcClient = &RoomGrpcClient{}
	}
	return mRoomGrpcClient
}

func (this *RoomGrpcClient) InitRetry() bool {
	var err error
	for i := 0; i < 10; i++ {
		this.conn, err = grpc.Dial(env.Get("room", "grpc_server"), grpc.WithInsecure())
		if err != nil {
			glog.Errorln("[RoomGrpcClient] connect failed:", err)
			time.Sleep(2 * time.Second)
			continue
		} else {
			break
		}
	}
	if this.conn == nil {
		return false
	}
	if !this.InitGrpcClient() {
		glog.Errorln("[RoomGrpcClient] init grpc client error:", err)
		return false
	}
	this.SendRegist()
	// 定时发送服务器信息给rcenterserver
	go this.TickerSendLoadInfo()
	return true
}

func (this *RoomGrpcClient) Init() bool {
	var err error
	this.conn, err = grpc.Dial(env.Get("room", "grpc_server"), grpc.WithInsecure())
	if err != nil {
		glog.Errorln("[RoomGrpcClient] connect failed:", err)
		return false
	}
	if !this.InitGrpcClient() {
		glog.Errorln("[RoomGrpcClient] init grpc client error:", err)
		return false
	}
	this.SendRegist()
	// 定时发送服务器信息给rcenterserver
	go this.TickerSendLoadInfo()

	return true
}

func (this *RoomGrpcClient) InitGrpcClient() bool {
	var err error
	client := usercmd.NewStreamRoomServiceClient(this.conn)
	this.mRouteClient, err = client.Route(context.Background())

	if err != nil {
		glog.Errorln("[InitGrpcClient] error:", err)
		return false
	}

	return true
}

// roomserver启动时主动向rcenterserver发送基本信息
func (this *RoomGrpcClient) SendRegist() bool {
	if this.mRouteClient == nil {
		glog.Errorln("[RoomGrpcClient] route client is nil")
		return false
	}

	p, _ := strconv.Atoi(*port)
	bytes, err := json.Marshal(struct {
		Ip   string `json:"ip"`
		Port int    `json:"port"`
	}{
		"localhost",
		p,
	})
	if err != nil {
		glog.Errorln("[RoomGrpcClient] struct to json error: ", err)
		return false
	}
	this.mRouteClient.Send(&usercmd.RoomRequest{
		Type: usercmd.RoomServerMsgType_Register,
		Data: bytes,
	})
	return true
}

// roomserver向rcenterserver发送负载信息（例如有多少房间、多少玩家等）
func (this *RoomGrpcClient) TickerSendLoadInfo() {
	if this.mRouteClient == nil {
		glog.Errorln("[RoomGrpcClient] route client is nil")
		return
	}
	info := &usercmd.RoomServerInfo{}
	// 固定信息
	p, _ := strconv.Atoi(*port)
	info.Ip = "localhost"
	info.Port = uint32(p)

	r := RoomManager_GetMe()

	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C
		// 获取实时信息
		info.RoomNum = r.curNum
		info.PlayerNum = r.curNum * 2 // 粗略计算（待优化）
		info.CurRoomId = r.curNum
		//
		bytes, err := json.Marshal(info)
		if err != nil {
			glog.Errorln("[RoomGrpcClient TickerSendLoadInfo] struct to json error: ", err)
			continue
		}
		this.mRouteClient.Send(&usercmd.RoomRequest{
			Type: usercmd.RoomServerMsgType_Update,
			Data: bytes,
		})
	}
}

// 房间服务器结束后，向rcenterserver发送信息
func (this *RoomGrpcClient) RemoveRoomServer() bool {
	if this.mRouteClient == nil {
		glog.Errorln("[RoomGrpcClient] route client is nil")
		return false
	}
	p, _ := strconv.Atoi(*port)
	bytes, err := json.Marshal(struct {
		Ip   string `json:"ip"`
		Port int    `json:"port"`
	}{
		"localhost",
		p,
	})
	if err != nil {
		glog.Errorln("[RoomGrpcClient] struct to json error: ", err)
		return false
	}
	this.mRouteClient.Send(&usercmd.RoomRequest{
		Type: usercmd.RoomServerMsgType_Remove,
		Data: bytes,
	})
	return true
}
