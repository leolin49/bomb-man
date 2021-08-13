package common

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"io"
	"paopao/server/usercmd"

	"github.com/golang/glog"
)

const (
	MaxCompressSize = 1024
	CmdHeaderSize   = 2
)

type Message interface {
	Marshal() (data []byte, err error)
	MarshalTo(data []byte) (n int, err error)
	Size() (n int)
	Unmarshal(data []byte) error
}

// 压缩
func zlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	_, err := w.Write(src)
	if err != nil {
		return nil
	}
	w.Close()
	return in.Bytes()
}

// 解压缩
func zlibUnCompress(src []byte) []byte {
	b := bytes.NewReader(src)
	var out bytes.Buffer
	r, err := zlib.NewReader(b)
	if err != nil {
		return nil
	}
	_, err = io.Copy(&out, r)
	if err != nil {
		return nil
	}
	return out.Bytes()
}

// 生成二进制数据,返回数据和是否压缩标识
func EncodeCmd(cmd uint16, msg Message) ([]byte, byte, error) {
	msglen := msg.Size()
	if msglen >= MaxCompressSize {
		data, err := msg.Marshal()
		if err != nil {
			glog.Errorln("[协议] 编码错误 ", err)
			return nil, 0, err
		}
		mbuff := zlibCompress(data)
		p := make([]byte, CmdHeaderSize+len(mbuff))
		p[0] = byte(cmd)
		p[1] = byte(cmd >> 8)
		copy(p[CmdHeaderSize:], mbuff)
		return p, 1, nil
	}
	p := make([]byte, CmdHeaderSize+msglen)
	_, err := msg.MarshalTo(p[CmdHeaderSize:])
	if err != nil {
		glog.Errorln("[协议] 编码错误 ", err)
		return nil, 0, err
	}
	p[0] = byte(cmd)
	p[1] = byte(cmd >> 8)
	return p, 0, nil
}

// 生成二进制数据,返回数据和是否压缩标识(json)
func EncodeCmdByJson(cmd uint16, msg Message) ([]byte, byte, error) {

	txt, _ := json.Marshal(msg)
	info := usercmd.CmdHeader{
		Cmd:  usercmd.MsgTypeCmd(cmd),
		Data: string(txt),
	}
	res, _ := json.Marshal(&info)

	return res, 0, nil
}

func EncodeToBytes(cmd uint16, msg Message) ([]byte, bool) {
	data := make([]byte, CmdHeaderSize+msg.Size())
	_, err := msg.MarshalTo(data[CmdHeaderSize:])
	if err != nil {
		glog.Errorln("[协议] 编码错误", err)
		return nil, false
	}
	data[0] = byte(cmd)
	data[1] = byte(cmd >> 8)
	return data, true
}

func EncodeToBytesJson(cmd uint16, msg Message) ([]byte, bool) {

	txt, _ := json.Marshal(msg)
	info := usercmd.CmdHeader{
		Cmd:  usercmd.MsgTypeCmd(cmd),
		Data: string(txt),
	}
	res, _ := json.Marshal(&info)

	return res, true
}

// protobuf解码数据
func DecodeGoMsg(buf []byte, flag byte, pb Message) (err error) {
	var mbuff []byte
	if flag == 1 {
		mbuff = zlibUnCompress(buf)
	} else {
		mbuff = buf
	}
	err = pb.Unmarshal(mbuff)
	if err != nil {
		glog.Errorln("[协议] gogo解码错误 ", err)
	}
	return
}

// protobuf解码数据
func DecodeGoCmd(buf []byte, flag byte, pb Message) (err error) {
	if len(buf) < CmdHeaderSize {
		glog.Errorln("[协议] 解码错误，长度过短")
		return
	}
	err = DecodeGoMsg(buf[CmdHeaderSize:], flag, pb)
	return
}

// 获取指令号
func GetCmd(buf []byte) uint16 {
	if len(buf) < CmdHeaderSize {
		return 0
	}
	return uint16(buf[0]) | uint16(buf[1])<<8
}

// 生成protobuf数据
func DecodeCmd(buf []byte, flag byte, pb Message) Message {
	if len(buf) < CmdHeaderSize {
		glog.Errorln("[协议] 数据错误 ", buf)
		return nil
	}
	var mbuff []byte
	if flag == 1 {
		mbuff = zlibUnCompress(buf[CmdHeaderSize:])
	} else {
		mbuff = buf[CmdHeaderSize:]
	}
	err := pb.Unmarshal(mbuff)
	if err != nil {
		glog.Errorln("[协议] 解码错误 ", err, ",", mbuff)
		return nil
	}
	return pb
}
