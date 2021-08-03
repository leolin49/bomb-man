package main

import (
	"net"
	"paopao/server-base/src/base/env"

	"github.com/golang/glog"
)

func HeartBeatInit() {
	server := env.Get("logic", "heartbeat_server")
	lis, err := net.Listen("tcp", server)
	if err != nil {
		glog.Errorln("[HeartBeat] listen error: ", err)
		return
	}
	for {
		conn, err := lis.Accept()
		if err != nil {
			glog.Errorln("[HeartBeat] accept error: ", err)
			continue
		}
		glog.Infoln("[HeartBeat] connect success: ", err)
		go HeartBeatCheck(conn)
	}
}

func HeartBeatCheck(conn net.Conn) {
	// TODO:心跳检查
	// buffer := make([]byte, 1024)
	// for {
	// 	n, err := conn.Read(buffer)

	// }
}
