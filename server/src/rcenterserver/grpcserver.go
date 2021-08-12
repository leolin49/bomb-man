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
		case usercmd.RoomServerMsgType_Register:
			glog.Infoln("[RoomGrpcServer] get one <RoomServerMsgType_Register> message")
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
			RoomServerManager_GetMe().AddRoomServer(info.Ip, info.Port)
			break
		case usercmd.RoomServerMsgType_Update:
			glog.Infoln("[RoomGrpcServer] get one <RoomServerMsgType_Update> message")
			info := &usercmd.RoomServerInfo{}
			err := json.Unmarshal(stream.Data, info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			RoomServerManager_GetMe().UpdateRoomServer(info)
			break
		case usercmd.RoomServerMsgType_Remove:
			glog.Infoln("[RoomGrpcServer] get one <RoomServerMsgType_Remove> message")
			var info struct {
				Ip   string `json:"ip"`
				Port uint32 `json:"port"`
			}
			err := json.Unmarshal(stream.Data, &info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			glog.Infof("[删除房间服务器] ip: %v, port: %v", info.Ip, info.Port)
			RoomServerManager_GetMe().DeleteRoomServer(info.Ip, info.Port)
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
