package main

import (
	"math/rand"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

const (
	RoomLoginToken = "FegiEcsOLtnSCSNp/lzofyl3tRGyl6asHEvYBAlMnv1MKgyE/r53/nInU0ae0yu8ay2T+6o5xztgtGIW38Q07Cs2/uNi79GIb0Po+/yz8NDInV9EvJm6PGqQlT4+TojxYGrsp+DTuYbeqf6txwRUmQWHhN8Eg0x6aAT4rmW53Mk="
)

type Robot struct {
	id      uint64           // 玩家id
	name    string           // 玩家昵称
	key     string           // 登录token
	score   uint32           // 分数
	curPos  *common.Position // 当前位置
	nextPos *common.Position // 下一个位置
	hp      int32            // 生命值
	curbomb uint32           // 当前已经放置的炸弹数量
	maxbomb uint32           // 能放置的最大炸弹数量
	power   uint32           // 炸弹威力
	speed   float64          // 移动速度

	client *LogicClient
}

var robot *Robot

func Robot_GetMe() *Robot {
	if robot == nil {
		robot = &Robot{
			id:      100005,
			name:    "abc",
			score:   0,
			hp:      3,
			curbomb: 0,
			maxbomb: 5,
			power:   1,
			speed:   0.2,
		}
	}
	return robot
}

func (this *Robot) Connect(addr string) bool {
	this.client = NewClient()
	return this.client.Connect(addr)
}

func (this *Robot) Move(way int32) {
	retCmd := &usercmd.MsgMove{
		Way: way,
	}
	this.client.SendCmd(usercmd.MsgTypeCmd_Move, retCmd)
}

func (this *Robot) PutBomb() {
	retCmd := &usercmd.MsgPutBomb{
		None: true,
	}
	this.client.SendCmd(usercmd.MsgTypeCmd_PutBomb, retCmd)
}

func (this *Robot) StartRandOperate() {
	for {
		x := rand.Int31() % 2
		switch x {
		case 0:
			w := rand.Int31() % 4
			this.Move(w)
			glog.Infoln("[move] way:", w)
		case 1:
			this.PutBomb()
			glog.Infoln("[put bomb]")
		}
		time.Sleep(time.Second * 2)
	}
}

// 发送房间验证信息
func (this *Robot) SendRoomToken(token string) {
	info := &usercmd.UserLoginInfo{
		Token: token,
	}
	this.client.SendCmd(usercmd.MsgTypeCmd_Login, info)
}

func main() {

	rand.Seed(time.Now().Unix())

	robot := Robot_GetMe()
	if !robot.Connect("127.0.0.1:13000") {
		glog.Errorln("[无法连接房间服务器]")
		return
	}
	robot.SendRoomToken(RoomLoginToken)

	go robot.StartRandOperate()

	for {

	}
}
