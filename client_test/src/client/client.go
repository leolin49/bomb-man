package main

import (
	"fmt"
	"paopao/server-base/src/base/gonet"
	"paopao/server/src/common"
	"paopao/server/usercmd"

	"github.com/golang/glog"
)

type LogicClient struct {
	gonet.TcpTask
	mclient *gonet.TcpClient
}

func NewClient() *LogicClient {
	s := &LogicClient{
		TcpTask: *gonet.NewTcpTask(nil),
	}
	s.Derived = s
	return s
}

func (this *LogicClient) Connect(addr string) bool {
	conn, err := this.mclient.Connect(addr)
	if err != nil {
		fmt.Println("连接失败 ", addr)
		return false
	}

	this.Conn = conn
	this.Start()

	fmt.Println("连接成功 ", addr)
	return true
}

func (this *LogicClient) ParseMsg(data []byte, flag byte) bool {

	cmd := usercmd.MsgTypeCmd(common.GetCmd(data))

	switch cmd {
	case usercmd.MsgTypeCmd_SceneSync:
		revCmd := &usercmd.RetUpdateSceneMsg{}
		if common.DecodeGoCmd(data, flag, revCmd) != nil {
			return false
		}
		glog.Infoln("===============[收到场景同步信息]===============")
		// 玩家信息
		fmt.Println("--------------[玩家信息]-------------")
		for _, v := range revCmd.Players {
			info := fmt.Sprintf("[player]id:%v, x:%v, y:%v, bombnum:%v, hp:%v",
				v.Id, v.X, v.Y, v.BombNum, v.State)
			fmt.Println(info)
		}
		// 炸弹信息
		fmt.Println("--------------[炸弹信息]-------------")
		for _, v := range revCmd.Bombs {
			info := fmt.Sprintf("[bomb]id:%v, x:%v, y:%v, own:%v",
				v.Id, v.X, v.Y, v.Own)
			fmt.Println(info)
		}
	}
	return true
}

func (this *LogicClient) SendCmd(cmd usercmd.MsgTypeCmd, msg common.Message) bool {
	data, flag, err := common.EncodeCmd(uint16(cmd), msg)
	if err != nil {
		fmt.Println("[服务] 发送失败 cmd:", cmd, ",len:", len(data), ",err:", err)
		return false
	}
	return this.AsyncSend(data, flag)
}

func (this *LogicClient) OnClose() {

}
