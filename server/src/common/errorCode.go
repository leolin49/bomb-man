package common

const (
	ErrorCodeSuccess        = 1000 // 成功
	ErrorCodeUserNotExist   = 2001 // 登录失败，用户不存在
	ErrorCodePassWordWrong  = 2002 // 登录失败，密码错误
	ErrorCodeUserNameRepeat = 2003 // 注册失败，用户名已存在
	ErrorCodeInvalidToken   = 2004 // 无效的token（token错误或已过期）
	ErrorCodeRoom           = 3001 // 进入房间错误
	ErrorCodeServer         = 5000 // 服务器内部错误
	ErrorUnknow             = 6000 // 未知错误
)
