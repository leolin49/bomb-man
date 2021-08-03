package gonet

import (
	"container/list"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

const (
	CMD_HEADER_SIZE = 4
	CMD_VERIFY_TIME = 30
	Cmd_MAX_SIZE    = 512 * 1024
)

type ITcpTask interface {
	ParseMessage(data []byte, flag byte) bool
	OnClose()
}

type TcpTask struct {
	closed      int32
	verified    bool // 客户端验证
	stopedChan  chan bool
	sendMsgList *list.List
	sendMutex   sync.Mutex
	Conn        *websocket.Conn
	Derived     ITcpTask
	msgChan     chan []byte
}

func NewTcpTask(conn *websocket.Conn) *TcpTask {
	return &TcpTask{
		closed:      -1,
		verified:    false,
		Conn:        conn,
		stopedChan:  make(chan bool, 1), // 容量为1的channel有缓冲
		sendMsgList: list.New(),
		msgChan:     make(chan []byte, 1024),
	}
}

func (this *TcpTask) Start() {
	// CompareAndSwapInt32: 判断参数addr指向的值是否与参数old的值相等，
	// 如果相等，用参数new的新值替换掉addr存储的旧值，否则操作就会被忽略。交换成功，返回true.
	if atomic.CompareAndSwapInt32(&this.closed, -1, 0) {
		glog.Info("[WebSocketTask Connect] Got Connect, ", this.Conn.RemoteAddr())
		go this.sendloop()
		go this.recvloop()
	}
}

func (this *TcpTask) Stop() bool {
	if !this.IsClosed() && len(this.stopedChan) == 0 {
		this.stopedChan <- true
	} else {
		glog.Info("[WebSocketTask Connect] Stop Connect Fail ", len(this.stopedChan))
		return false
	}
	glog.Info("[WebSocketTask Connect] Stop Connect Success")
	return true
}

func (this *TcpTask) Close() {
	if atomic.CompareAndSwapInt32(&this.closed, 0, 1) {
		this.Conn.Close()
		this.Derived.OnClose()
		close(this.stopedChan)

		glog.Info("[WebSocketTask Connect] Connect Close ", this.Conn.RemoteAddr())
	}
}

func (this *TcpTask) Reset() {
	if atomic.LoadInt32(&this.closed) == 1 {
		glog.Info("[WebSocketTask Connect] Connect Reset ", this.Conn.RemoteAddr())
		this.closed = -1
		this.verified = false
		this.stopedChan = make(chan bool)
	}
}

func (this *TcpTask) AsyncSend(buffer []byte, flag byte) bool {
	if this.IsClosed() {
		return false
	}

	bufsize := len(buffer)
	totalsize := bufsize + CMD_HEADER_SIZE

	// sendbuffer初始大小为0
	sendbuffer := make([]byte, 0, totalsize)
	// TODO
	// 00000101 00101000 10100100 10010101
	sendbuffer = append(sendbuffer, byte(bufsize), byte(bufsize>>8), byte(bufsize>>16), flag)
	sendbuffer = append(sendbuffer, buffer...)
	this.msgChan <- sendbuffer

	return true
}

func (this *TcpTask) recvloop() {
	defer func() {
		if err := recover(); err != nil {
			glog.Error("[WebSocketTask Unexpeted] ", err, "\n", string(debug.Stack()))
		}
	}()
	defer this.Close()

	var datasize int

	for !this.IsClosed() {
		_, bytemsg, err := this.Conn.ReadMessage()
		if nil != err {
			glog.Error("[WebSocketTask Recv] Recv Failed ", this.Conn.RemoteAddr(), ",", err)
			return
		}
		// 根据报文头部信息，计算有效数据长度
		datasize = int(bytemsg[0]) | int(bytemsg[1])<<8 | int(bytemsg[2])<<16
		if datasize > Cmd_MAX_SIZE {
			glog.Error("[WebSocketTask Recv] Package Too Large ", this.Conn.RemoteAddr(), ",", datasize)
			return
		}
		// 解析报文
		this.Derived.ParseMessage(bytemsg[CMD_HEADER_SIZE:], bytemsg[3])
	}
}

func (this *TcpTask) sendloop() {
	defer func() {
		// Recover 是一个Go语言的内建函数，可以让进入宕机流程中的 goroutine 恢复过来，recover 仅在延迟函数 defer 中有效，在正常的执行过程中，调用 recover 会返回 nil 并且没有其他任何效果，
		// 如果当前的 goroutine 陷入恐慌，调用 recover 可以捕获到 panic 的输入值，并且恢复正常的执行。
		// 通常来说，不应该对进入 panic 宕机的程序做任何处理，但有时，需要我们可以从宕机中恢复，
		// 至少我们可以在程序崩溃前，做一些操作，举个例子，当 web 服务器遇到不可预料的严重问题时，在崩溃前应该将所有的连接关闭，
		// 如果不做任何处理，会使得客户端一直处于等待状态，如果 web 服务器还在开发阶段，服务器甚至可以将异常信息反馈到客户端，帮助调试。
		if err := recover(); err != nil {
			glog.Error("[WebSocketTask Unexpeted] ", err, "\n", string(debug.Stack()))
		}
	}()
	defer this.Close()

	// 超时时间
	var timeout *time.Timer
	timeout = time.NewTimer(time.Second * CMD_VERIFY_TIME)

	for {
		select {
		case bytemsg := <-this.msgChan:
			if nil != bytemsg && len(bytemsg) > 0 {
				err := this.Conn.WriteMessage(websocket.BinaryMessage, bytemsg)
				if nil != err {
					glog.Error("[WebSocketTask Send] Send Failed ", this.Conn.RemoteAddr(), ", ", err)
					return
				}
			} else {
				glog.Error("[WebSocketTask Send] Wrong Message ", bytemsg)
				return
			}
		case <-this.stopedChan:
			return
		case <-timeout.C:
			// 超时，防止用户连接而不使用，浪费服务器资源
			if !this.IsVerifed() {
				glog.Error("[WebSocketTask] Client Verify Timeout ", this.Conn.RemoteAddr())
			}
		}
	}
}

func (this *TcpTask) IsClosed() bool {
	return atomic.LoadInt32(&this.closed) != 0
}

func (this *TcpTask) Verify() {
	this.verified = true
}

func (this *TcpTask) IsVerifed() bool {
	return this.verified
}

func (this *TcpTask) Terminate() {
	this.Close()
}
