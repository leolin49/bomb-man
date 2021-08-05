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
	USER_ID_BASENUM     = 100000 // 用户ID基数
	TOKEN_EXPIRE_TIME_S = 300    // token过期时间（s）
)

const ( // http状态码
	Unauthorized          = 401 // 请求要求用户的身份认证
	Not_Found             = 404 // 服务器无法根据客户端的请求找到资源
	Internal_Server_Error = 500 // 服务器内部错误，无法完成请求
)

const ( //
	SUCCESS                         = 1000 // 成功
	LOGIN_ERROR_USERNOTEXSIT        = 2001 // 登录失败，用户不存在
	LOGIN_ERROR_PASSWORDERROR       = 2002 // 登录失败，密码错误
	REGISTER_ERROR_USERALREADYEXSIT = 2003 // 注册失败，用户名已存在
	INVALID_TOKEN                   = 2004 // 无效的token（token错误或已过期）

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
	}{SUCCESS})
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
		glog.Infoln("[User Register] username existed")
		writer.Write(JustRetCodeJson(REGISTER_ERROR_USERALREADYEXSIT))
		return
	}
	count++
	uid := USER_ID_BASENUM + count
	common.RedisMgr.HMSet(key, UserInfo{
		Id:           strconv.Itoa(uid),
		Password:     pwd,
		Registertime: curtime,
	})
	// 更新用户数量
	common.RedisMgr.Set("PlayerNumber", strconv.Itoa(count))
	glog.Infof("[Register success]%v, %v, %v, %v", key, uid, pwd, curtime)
	// 注册成功
	writer.Write(JustRetCodeJson(SUCCESS))
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
		writer.Write(JustRetCodeJson(LOGIN_ERROR_USERNOTEXSIT))
		return
	}
	// 验证2：密码错误
	if pwd != common.RedisMgr.HGet(key, "PassWord") {
		glog.Infoln("[User Login] password error ", pwd)
		writer.Write(JustRetCodeJson(LOGIN_ERROR_PASSWORDERROR))
		return
	}
	// 生成token
	token, err := common.CreateLoginToken(key)
	if len(token) == 0 || err != nil {
		glog.Errorln("[User Login] create token error", err)
		return
	}
	// token存入数据库，设置过期时间
	tokenKey := key + "_token"
	common.RedisMgr.Set(tokenKey, token)
	conn := common.RedisMgr.RedisPool.Get()
	defer conn.Close()
	conn.Do("EXPIRE", tokenKey, TOKEN_EXPIRE_TIME_S)
	glog.Infoln("[User Login] login success, username:", key)
	// 登录成功
	tmp := struct {
		ResultCode int    `json:"result_code"`
		Token      string `json:"token"`
	}{
		SUCCESS,
		token,
	}
	res, _ := json.Marshal(tmp)
	writer.Write(res)
}

// 开始游戏请求，将玩家信息通过rpc给到rcenterserver
func StartGameHandler(writer http.ResponseWriter, request *http.Request) {
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
		writer.Write(JustRetCodeJson(INVALID_TOKEN))
		writer.WriteHeader(Unauthorized)
		return
	}

	id, err := strconv.ParseUint(common.RedisMgr.HGet(username, "Id"), 10, 64)
	if err != nil {
		glog.Errorln("[Start Game] uid empty, username:", username)
		return
	}

	// TODO:rpc请求，获取房间信息
	reqData := usercmd.ReqIntoRoom{
		UId:      &id,
		UserName: &username,
	}
	rspData := RequestRpcService(&reqData)
	res, _ := json.Marshal(rspData)

	// 将房间信息返回给用户
	writer.Write(res)
}
