package main

import (
	"encoding/json"
	"io"
	"net"
	"paopao/server-base/src/base/env"
	"paopao/server/usercmd"

	"github.com/golang/glog"
	"google.golang.org/grpc"
)

type RoomGrpcService struct {
}

func (this *RoomGrpcService) Route(conn usercmd.StreamRoomService_RouteServer) error {
	for {
		stream, err := conn.Recv()
		if err != nil {
			if err == io.EOF {
				glog.Infoln("[RoomGrpcServer] got EOF")
				return nil
			} else {
				glog.Errorln("[RoomGrpcServer] server error: ", err)
				return err
			}
		}

		switch stream.Type {
		case usercmd.RoomMsgType_RegisterRoom:
			glog.Infoln("[RoomGrpcServer] get one <RoomMsgType_RegisterRoom> message")
			var info struct {
				Ip   string `json:"ip"`
				Port uint32 `json:"port"`
			}
			err := json.Unmarshal(stream.Data, &info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			glog.Infof("[注册房间服务器] ip: %v, port: %v", info.Ip, info.Port)
			RCenterServer_GetMe().AddRoomServer(info.Ip, info.Port)
			break
		case usercmd.RoomMsgType_UpdateRoom:
			glog.Infoln("[RoomGrpcServer] get one <RoomMsgType_UpdateRoom> message")
			info := &usercmd.RoomServerInfo{}
			err := json.Unmarshal(stream.Data, info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			RCenterServer_GetMe().UpdateRoomServer(info)
			break
		}
	}
}

func GrpcServerStart() bool {
	grpcServer := grpc.NewServer()
	usercmd.RegisterStreamRoomServiceServer(grpcServer, new(RoomGrpcService))
	glog.Infoln("[GrpcServerStart] address: ", env.Get("rcenter", "grpc_server"))
	listener, err := net.Listen("tcp", env.Get("rcenter", "grpc_server"))
	if err != nil {
		glog.Errorln("[GrpcServerStart] grpc service start error:", err)
		return false
	}
	go func() {
		grpcServer.Serve(listener)
	}()
	glog.Infoln("[GrpcServerStart] grpc service start success")
	return true
}
