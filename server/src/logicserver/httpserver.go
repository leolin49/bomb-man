package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"paopao/server/src/common"
	"paopao/server/usercmd"
	"strconv"
	"time"

	"github.com/golang/glog"
)

const (
	UserIdBaseNumber = 100000 // 用户ID基数
	TokenExpireTime  = 300    // token过期时间（s）
)

const ( // http状态码
	Unauthorized          = 401 // 请求要求用户的身份认证
	Not_Found             = 404 // 服务器无法根据客户端的请求找到资源
	Internal_Server_Error = 500 // 服务器内部错误，无法完成请求
)

type UserInfo struct {
	Id           string `redis:"Id"`           // 玩家ID
	Password     string `redis:"PassWord"`     // 密码
	Registertime string `redis:"RegisterTime"` // 注册时间
}

// 用于只带有 操作结果码 的json响应体
func JustRetCodeJson(code int) []byte {
	res, _ := json.Marshal(struct {
		ResultCode int `json:"result_code"`
	}{code})
	return res
}

func InitHttpServer() bool {

	http.HandleFunc("/test", testHandler)

	http.HandleFunc("/register", RegisterHandler)
	http.HandleFunc("/login", LoginHandler)
	http.HandleFunc("/start", StartGameHandler)

	listen, err := net.Listen("tcp", "localhost:"+*port)
	if err != nil {
		glog.Errorln("[InitHttpServer] band port error")
		return false
	}
	glog.Infoln("[InitHttpServer] Listen port ", *port)

	ser := &http.Server{
		// IdleTimeout 是启用 keep-alives 时等待下一个请求的最长时间。
		// 如果 IdleTimeout 为零，则使用 ReadTimeout 的值。 如果两者都为零，则没有超时。
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 60 * time.Second,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
	}

	go ser.Serve(listen)

	glog.Infoln("[InitHttpServer] Service start success")
	return true
}

func testHandler(writer http.ResponseWriter, request *http.Request) {
	_, err := fmt.Fprintln(writer, "hello world by localhost"+*port)
	if err != nil {
		return
	}
}

// 注册账号
func RegisterHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("content-type", "text/json")
	request.ParseForm()
	key := request.Form.Get("username")
	pwd := request.Form.Get("password")
	curtime := GetCurrentTime()
	count, err := strconv.Atoi(common.RedisMgr.Get("PlayerNumber"))
	if err != nil { // key 'PlayerNumber' 不存在
		writer.WriteHeader(Internal_Server_Error)
		return
	}
	// 用户名已存在
	if common.RedisMgr.Exist(key) {
		glog.Infof("[User Register] username [%v] existed", key)
		writer.Write(JustRetCodeJson(common.ErrorCodeUserNameRepeat))
		return
	}
	count++
	uid := UserIdBaseNumber + count
	common.RedisMgr.HMSet(key, UserInfo{
		Id:           strconv.Itoa(uid),
		Password:     pwd,
		Registertime: curtime,
	})
	// 更新用户数量
	common.RedisMgr.Set("PlayerNumber", strconv.Itoa(count))
	glog.Infof("[Register success]%v, %v, %v, %v", key, uid, pwd, curtime)
	// 注册成功
	writer.Write(JustRetCodeJson(common.ErrorCodeSuccess))
}

// 登录游戏
func LoginHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("content-type", "text/json")
	request.ParseForm()
	key := request.Form.Get("username")
	pwd := request.Form.Get("password")
	// 验证1：用户名不存在
	if !common.RedisMgr.Exist(key) {
		glog.Infoln("[User Login] user not exist ", key)
		writer.Write(JustRetCodeJson(common.ErrorCodeUserNotExist))
		return
	}
	// 验证2：密码错误
	if pwd != common.RedisMgr.HGet(key, "PassWord") {
		glog.Infoln("[User Login] password error ", pwd)
		writer.Write(JustRetCodeJson(common.ErrorCodePassWordWrong))
		return
	}
	// 生成token
	token, err := common.CreateLoginToken(key)
	if len(token) == 0 || err != nil {
		glog.Errorln("[User Login] create token error", err)
		return
	}
	// token存入数据库，设置过期时间
	tokenKey := key + "_logintoken"
	common.RedisMgr.Set(tokenKey, token)
	conn := common.RedisMgr.RedisPool.Get()
	defer conn.Close()
	conn.Do("EXPIRE", tokenKey, TokenExpireTime)
	glog.Infoln("[User Login] login success, username:", key)
	// 登录成功
	tmp := struct {
		ResultCode int    `json:"result_code"`
		Token      string `json:"login_token"`
	}{
		common.ErrorCodeSuccess,
		token,
	}
	res, _ := json.Marshal(tmp)
	writer.Write(res)
}

func StartGameHandler(writer http.ResponseWriter, request *http.Request) {
	req := struct {
		Token string `json:"login_token"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		body, _ := ioutil.ReadAll(request.Body)
		glog.Errorln("[Start Game] json parse error: ", string(body))
		return
	}
	// 解析token
	username, err := common.ParseLoginToken(req.Token)
	if err != nil {
		glog.Errorln("[Start Game] ", err)
		writer.Write(JustRetCodeJson(common.ErrorUnknow))
		return
	}
	// 验证token是否正确，是否过期
	if t := common.RedisMgr.Get(username + "_logintoken"); t != req.Token {
		glog.Errorln("[Start Game] token error or expired")
		glog.Errorln("[Start Game] redis token: ", t)
		glog.Errorln("[Start Game] request token: ", req.Token)
		writer.Write(JustRetCodeJson(common.ErrorCodeInvalidToken))
		writer.WriteHeader(Unauthorized)
		return
	}

	id, err := strconv.ParseUint(common.RedisMgr.HGet(username, "Id"), 10, 64)
	if err != nil {
		glog.Errorln("[Start Game] uid empty, username:", username)
		writer.Write(JustRetCodeJson(common.ErrorUnknow))
		return
	}

	// rpc请求，获取房间信息
	reqData := usercmd.ReqIntoRoom{
		UId:      &id,
		UserName: &username,
	}
	var rspData *usercmd.RetIntoRoom
	rspData = RequestRpcService(&reqData)

	// 根据返回的roomserver_addr和roomid生成roomtoken，用于客户端连接roomserver
	roomid := *rspData.RoomId
	token, err := common.CreateRoomToken(common.RoomTokenInfo{
		UserId:   id,
		UserName: username,
		RoomId:   roomid,
	})
	// roomtoken存入redis，key:username_roomtoken
	tokenKey := fmt.Sprintf("%v_roomtoken", username)
	common.RedisMgr.Set(tokenKey, token)
	conn := common.RedisMgr.RedisPool.Get()
	defer conn.Close()
	conn.Do("EXPIRE", tokenKey, TokenExpireTime)

	if err != nil {
		glog.Errorln("[HttpServer] 生成roomtoken失败, ", err)
		writer.Write(JustRetCodeJson(common.ErrorCodeServer))
		return
	}
	rspData.Key = &token

	res, _ := json.Marshal(struct {
		ResultCode int    `json:"result_code"`
		Token      string `json:"room_token"`
		Address    string `json:"room_address"`
	}{
		common.ErrorCodeSuccess,
		token,
		*rspData.Addr,
	})
	glog.Infoln("[用户请求开始游戏] username:", username)
	// 将房间信息返回给用户
	writer.Write(res)
}

// 开始游戏请求，将玩家信息通过rpc给到rcenterserver (token在recenterserver生成)
func StartGameHandler_bak(writer http.ResponseWriter, request *http.Request) {
	req := struct {
		Token string `json:"token"`
	}{}
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		body, _ := ioutil.ReadAll(request.Body)
		glog.Errorln("[Start Game] json parse error: ", string(body))
		return
	}
	// 解析token
	username, err := common.ParseLoginToken(req.Token)
	if err != nil {
		glog.Errorln("[Start Game] ", err)
		return
	}
	// 验证token是否正确，是否过期
	if t := common.RedisMgr.Get(username + "_token"); t != req.Token {
		glog.Errorln("[Start Game] token error or expired")
		glog.Errorln("[Start Game] redis token: ", t)
		glog.Errorln("[Start Game] request token: ", req.Token)
		writer.Write(JustRetCodeJson(common.ErrorCodeInvalidToken))
		writer.WriteHeader(Unauthorized)
		return
	}

	id, err := strconv.ParseUint(common.RedisMgr.HGet(username, "Id"), 10, 64)
	if err != nil {
		glog.Errorln("[Start Game] uid empty, username:", username)
		return
	}

	// rpc请求，获取房间信息
	reqData := usercmd.ReqIntoRoom{
		UId:      &id,
		UserName: &username,
	}
	rspData := RequestRpcService(&reqData)
	res, _ := json.Marshal(rspData)

	// 将房间信息返回给用户
	writer.Write(res)
}
