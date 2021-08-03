package gonet

import (
	"net"
	"time"

	"github.com/golang/glog"
)

const (
	READ_BUFFER_SIZE  = 128 * 1024
	WRITE_BUFFER_SIZE = 128 * 1024
)

type TcpServer struct {
	listener *net.TCPListener
}

func (this *TcpServer) Bind(address string) error {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if nil != err {
		glog.Error("[TcpServer Bind] Bind Failed ", address)
		return err
	}

	lis, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		glog.Error("[TcpServer Bind] Listen Failed ", address)
		return err
	}

	glog.Info("[TcpServer] Listen Success ", address)
	this.listener = lis
	return nil
}

func (this *TcpServer) Accept() (*net.TCPConn, error) {
	// SetDeadline 设置与侦听器关联的截止日期。 零时间值禁用最后期限。
	this.listener.SetDeadline(time.Now().Add(time.Second * 1))

	conn, err := this.listener.AcceptTCP()
	if err != nil {
		glog.Error("[TcpServer Accept] AcceptTCP Failed ")
		return nil, err
	}

	// 设置KeepAlive以及时间
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(1 * time.Minute)
	// SetNoDelay 控制操作系统是否应该延迟数据包传输以希望发送更少的数据包（Nagle 算法）。默认值为真（无延迟），意味着在写入后尽快发送数据。
	conn.SetNoDelay(true)
	// 设置缓冲区大小
	conn.SetWriteBuffer(WRITE_BUFFER_SIZE)
	conn.SetReadBuffer(READ_BUFFER_SIZE)

	return conn, nil
}
