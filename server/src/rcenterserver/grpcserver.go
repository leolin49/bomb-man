package main

import (
	"encoding/json"
	"io"
	"paopao/server/usercmd"

	"github.com/golang/glog"
)

type RoomGrpcServer struct {
}

func (this *RoomGrpcServer) Route(conn usercmd.StreamRoomService_RouteServer) error {
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
			var info struct {
				Ip   string `json:"ip"`
				Port int    `json:"port"`
			}
			err := json.Unmarshal(stream.Data, &info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			RCenterServer_GetMe().AddRoomServer(info.Ip, info.Port)
			break
		case usercmd.RoomMsgType_UpdateRoom:
			var info usercmd.RoomServerInfo
			err := json.Unmarshal(stream.Data, &info)
			if err != nil {
				glog.Errorln("[RoomGrpcServer] json to struct error: ", err)
				return err
			}
			RCenterServer_GetMe().UpdateRoomServer(info)
			break
		}
	}
}
