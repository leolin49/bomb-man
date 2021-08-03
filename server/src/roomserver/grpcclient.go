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

func (this *RoomGrpcClient) Init() bool {
	var err error
	this.conn, err = grpc.Dial(env.Get("room", "grpc_server"))
	if err != nil {
		glog.Errorln("[RoomGrpcClient] connect failed:", err)
		return false
	}
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
		Type: usercmd.RoomMsgType_RegisterRoom,
		Data: bytes,
	})
	return true
}

// roomserver向rcenterserver发送负载信息（例如有多少房间、多少玩家等）
func (this *RoomGrpcClient) TickerSendLoadInfo() {
	ticker := time.NewTicker(2 * time.Second)
	for {
		<-ticker.C
		this.mRouteClient.Send(&usercmd.RoomRequest{
			Type: usercmd.RoomMsgType_UpdateRoom,
			// TODO
			Data: []byte("todo"),
		})
	}
}
