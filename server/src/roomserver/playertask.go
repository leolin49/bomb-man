package main

import (
	"net"
	"paopao/server-base/src/base/gonet"
	"time"
)

type PlayerTask struct {
	gonet.TcpTask
	// udptask *snet.Session
	isUdp bool

	key  string
	id   uint64
	name string
	room *Room
	// udata *common.UserData
	uobjs []uint32

	power   uint32
	speed   uint32
	lifenum uint32
	state   uint32

	activeTime time.Time
	onlineTime int64
}

func NewPlayerTask(conn net.Conn) *PlayerTask {
	m := &PlayerTask{
		TcpTask:    *gonet.NewTcpTask(conn),
		activeTime: time.Now(),
	}
	m.Derived = m
	return m
}

func (this *PlayerTask) OnClose() {
	this.room = nil
}

func (this *PlayerTask) ParseMsg(data []byte, flag byte) bool {
	return true
}
