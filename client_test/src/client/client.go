package main

import (
	"encoding/json"
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

	// cmd := usercmd.MsgTypeCmd(common.GetCmd(data))
	info := usercmd.CmdHeader{}
	err := json.Unmarshal(data, &info)
	if err != nil {
		glog.Errorln("[json解析失败] ", err)
	}
	cmd := info.Cmd

	switch cmd {
	case usercmd.MsgTypeCmd_SceneSync:
		revCmd := &usercmd.RetUpdateSceneMsg{}
		json.Unmarshal([]byte(info.Data), revCmd)
		// if common.DecodeGoCmd(data, flag, revCmd) != nil {
		// 	return false
		// }
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
	case usercmd.MsgTypeCmd_Death:
		revCmd := &usercmd.RetRoleDeath{}
		json.Unmarshal([]byte(info.Data), revCmd)
		// if common.DecodeGoCmd(data, flag, revCmd) != nil {
		// 	return false
		// }
		glog.Infoln("===============[收到玩家死亡信息]===============")
		fmt.Printf("killerid:%v\nkillername:%v\nsorce:%v\n",
			revCmd.KillId, revCmd.KillName, revCmd.Score)
		// this.Close()
	case usercmd.MsgTypeCmd_EndRoom:
		revCmd := &usercmd.SettleMentInfo{}
		json.Unmarshal([]byte(info.Data), revCmd)
		glog.Infoln("===============[收到房间结算信息]===============")
		for _, p := range revCmd.Players {
			fmt.Printf("[player]id:%v, score:%v", p.Id, p.Score)
		}
		this.Close()
	}
	return true
}

func (this *LogicClient) SendCmd(cmd usercmd.MsgTypeCmd, msg common.Message) bool {
	data, flag, err := common.EncodeCmdByJson(uint16(cmd), msg)
	if err != nil {
		fmt.Println("[服务] 发送失败 cmd:", cmd, ",len:", len(data), ",err:", err)
		return false
	}
	return this.AsyncSend(data, flag)
}

func (this *LogicClient) OnClose() {

}
