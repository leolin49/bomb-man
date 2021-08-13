package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"time"

	"github.com/golang/glog"
)

const (
	LogicServerUrl = "http://localhost:9000/"
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

func (this *Robot) RegisterAccount(username, password string) {
	rsp, err := http.PostForm(LogicServerUrl+"register", url.Values{
		"username": {username},
		"password": {password},
	})
	if err != nil {
		glog.Errorln("[http请求失败] ", err)
		return
	}
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		glog.Errorln("[http body 读取失败] ", err)
		return
	}
	glog.Infoln("[http respon body] ", string(body))
}

func (this *Robot) UserLogin(username, password string) string {
	rsp, err := http.PostForm(LogicServerUrl+"login", url.Values{
		"username": {username},
		"password": {password},
	})
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		glog.Errorln("[http body 读取失败] ", err)
		return ""
	}
	glog.Infoln("[http respon body] ", string(body))
	info := struct {
		Code  int    `json:"result_code"`
		Token string `json:"login_token"`
	}{}
	if err := json.Unmarshal(body, &info); err != nil {
		glog.Errorln("[http json 解析失败] ", err)
	}
	// if err := json.NewDecoder(rsp.Body).Decode(&info); err != nil {
	// 	glog.Errorln("[http json 解析失败] ", err)
	// }
	glog.Errorln(info)
	return info.Token
}

func (this *Robot) RequestStartGame(token string) (string, string) {
	rsp, err := http.PostForm(LogicServerUrl+"start", url.Values{
		"logintoken": {token},
	})
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		glog.Errorln("[http body 读取失败] ", err)
		return "", ""
	}
	glog.Infoln("[http respon body] ", string(body))
	info := struct {
		Code    int    `json:"result_code"`
		Token   string `json:"room_token"`
		Address string `json:"room_address"`
	}{}
	if err := json.Unmarshal(body, &info); err != nil {
		glog.Errorln("[http json 解析失败] ", err)
	}
	// if err := json.NewDecoder(rsp.Body).Decode(&info); err != nil {
	// 	glog.Errorln("[http json 解析失败] ", err)
	// }
	return info.Token, info.Address
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
		x := rand.Int31() % 7
		// fmt.Println("[随机数] x = ", x)
		switch x {
		case 0, 1, 2, 3, 4, 5:
			w := rand.Int31() % 4
			this.Move(w + 1)
			glog.Infoln("[move] way:", w+1)
			break
		case 6:
			this.PutBomb()
			glog.Infoln("[put bomb]")
			break
		}
		time.Sleep(time.Second * 1)
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

	flag.Parse()
	defer glog.Flush()

	robot := Robot_GetMe()

	var (
		username   string
		password   string
		logintoken string
		roomtoken  string
		address    string
	)

	var cmd int
	for {
		fmt.Print("--rigister 1\n--login 2\n--start game 3\n")
		fmt.Println("cmd:")
		fmt.Scanln(&cmd)
		switch cmd {
		case 1:
			fmt.Println("用户名：")
			fmt.Scanln(&username)
			fmt.Println("密码：")
			fmt.Scanln(&password)
			robot.RegisterAccount(username, password)
		case 2:
			fmt.Println("用户名：")
			fmt.Scanln(&username)
			fmt.Println("密码：")
			fmt.Scanln(&password)
			logintoken = robot.UserLogin(username, password)
		case 3:
			roomtoken, address = robot.RequestStartGame(logintoken)
			if !robot.Connect(address) {
				glog.Errorln("[无法连接房间服务器]")
				return
			}
			robot.SendRoomToken(roomtoken)
			time.Sleep(3 * time.Second)
			go robot.StartRandOperate()
			for {
			}
		}
	}
}
