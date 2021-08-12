package main

import (
	"fmt"
	"math/rand"
	"paopao/server-base/src/base/gonet"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"time"
)

type Client struct {
	gonet.TcpTask
	mclient *gonet.TcpClient
}

func NewClient() *Client {
	s := &Client{
		TcpTask: *gonet.NewTcpTask(nil),
	}
	s.Derived = s
	return s
}

func (this *Client) Connect(addr string) bool {
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

func (this *Client) ParseMsg(data []byte, flag byte) bool {
	this.Verify()
	this.AsyncSend(data, flag)
	return true
}

func (this *Client) SendCmd(cmd usercmd.MsgTypeCmd, msg common.Message) bool {
	data, flag, err := common.EncodeCmd(uint16(cmd), msg)
	if err != nil {
		fmt.Println("[服务] 发送失败 cmd:", cmd, ",len:", len(data), ",err:", err)
		return false
	}
	return this.AsyncSend(data, flag)
}

func (this *Client) OnClose() {

}

func main() {

	r := rand.New(rand.NewSource(time.Now().Unix()))

	client := NewClient()
	if !client.Connect("127.0.0.1:13000") {
		return
	}
	for {
		x := r.Int31() % 2
		switch x {
		case 0:
			retCmd := &usercmd.MsgMove{
				Way: r.Int31() % 4,
			}
			client.SendCmd(usercmd.MsgTypeCmd_Move, retCmd)
		case 1:
			retCmd := &usercmd.MsgPutBomb{
				None: true,
			}
			client.SendCmd(usercmd.MsgTypeCmd_PutBomb, retCmd)
		}
		time.Sleep(time.Second * 2)
	}
}
