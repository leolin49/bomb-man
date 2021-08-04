package main

import (
	"paopao/server-base/src/base/gonet"
	"time"

	"github.com/gorilla/websocket"
)

type PlayerTask struct {
	tcptask *gonet.WebSocketTask
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

func NewPlayerTask(conn *websocket.Conn) *PlayerTask {
	m := &PlayerTask{
		tcptask:    gonet.NewWebSocketTask(conn),
		activeTime: time.Now(),
	}
	m.tcptask.Derived = m
	return m
}

func (this *PlayerTask) OnClose() {
	this.tcptask.Stop()
	this.room = nil
}

func (this *PlayerTask) ParseMessage(data []byte, flag byte) bool {
	return true
}
