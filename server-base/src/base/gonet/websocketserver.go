package gonet

import (
	"net/http"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

type IWebSocketServer interface {
	OnWebAccept(conn *websocket.Conn)
}

type WebSocketServer struct {
	WebDerived IWebSocketServer
}

var upgrader = websocket.Upgrader{} // 使用默认配置

func (this *WebSocketServer) WebBind(addr string) error {
	http.HandleFunc("/", this.WebListen)
	err := http.ListenAndServe(addr, nil)
	if nil != err {
		glog.Error("[WebSocketServer] init failed", addr)
	}
	return nil
}

func (this *WebSocketServer) WebListen(writer http.ResponseWriter, request *http.Request) {
	//该函数用于拦截或放行跨域请求。函数返回值为bool类型，即true放行，false拦截
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	client, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		glog.Error("[WebSocketServer] failed to upgrade", request.RemoteAddr, err)
		return
	}

	glog.Info("[WebSocketServer] recv connect ", request.RemoteAddr, writer.Header().Get("Origin"), request.Header.Get("Origin"))

	this.WebDerived.OnWebAccept(client)
}

func (this *WebSocketServer) WebClose() error {
	return nil
}
