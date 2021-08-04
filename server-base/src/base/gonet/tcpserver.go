package gonet

import (
	"net"
	"time"
)

type TcpServer struct {
	listener *net.TCPListener
}

func (this *TcpServer) Bind(address string) error {

	tcpAddr, err := net.ResolveTCPAddr("tcp4", address)
	if nil != err {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if nil != err {
		return err
	}

	this.listener = listener
	return nil
}

func (this *TcpServer) BindAccept(address string, handler func(*net.TCPConn)) error {
	err := this.Bind(address)
	if err != nil {
		return err
	}
	go func() {
		for {
			conn, err := this.Accept()
			if err != nil {
				continue
			}
			handler(conn)
		}
	}()
	return nil
}

func (this *TcpServer) Accept() (*net.TCPConn, error) {
	// SetDeadline 设置与侦听器关联的截止日期。 零时间值禁用最后期限。
	this.listener.SetDeadline(time.Now().Add(time.Second * 1))

	conn, err := this.listener.AcceptTCP()
	if err != nil {
		return nil, err
	}

	// 设置KeepAlive以及时间
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(1 * time.Minute)
	// SetNoDelay 控制操作系统是否应该延迟数据包传输以希望发送更少的数据包（Nagle 算法）。默认值为真（无延迟），意味着在写入后尽快发送数据。
	conn.SetNoDelay(true)
	// 设置缓冲区大小
	conn.SetWriteBuffer(128 * 1024)
	conn.SetReadBuffer(128 * 1024)

	return conn, nil
}

func (this *TcpServer) Close() error {
	return this.listener.Close()
}
