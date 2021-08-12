package common

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"os"

	"github.com/golang/glog"
)

const (
	RSA_PUBLIC_FILEPATH  = "../../bin/rsapem/rsa_public_key.pem"
	RSA_PRIVATE_FILEPATH = "../../bin/rsapem/rsa_private_key.pem"
)

func ReadFileAll(filepath string) ([]byte, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func CreateLoginToken(username string) (string, error) {
	pubkey, err := ReadFileAll(RSA_PUBLIC_FILEPATH)
	if err != nil {
		glog.Errorln("[CreateToken] read file error")
		return "", err
	}
	token, err := RSAEncrypt([]byte(username), pubkey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

func ParseLoginToken(token string) (string, error) {
	prikey, err := ReadFileAll(RSA_PRIVATE_FILEPATH)
	if err != nil {
		glog.Errorln("[GetToken] read file error")
		return "", err
	}
	// base64
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		glog.Errorln("[GetToken]", err)
		return "", err
	}
	// rsa
	res, err := RSADecrypt(bytes, prikey)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func CreateRoomToken(info RoomTokenInfo) (string, error) {
	pubkey, err := ReadFileAll(RSA_PUBLIC_FILEPATH)
	if err != nil {
		glog.Errorln("[CreateToken] read file error")
		return "", err
	}
	bytes, err := json.Marshal(info)
	if err != nil {
		glog.Errorln("[CreateToken] json marshal error")
		return "", err
	}
	token, err := RSAEncrypt(bytes, pubkey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(token), nil
}

func ParseRoomToken(token string) (*RoomTokenInfo, error) {
	prikey, err := ReadFileAll(RSA_PRIVATE_FILEPATH)
	if err != nil {
		glog.Errorln("[GetToken] read file error")
		return nil, err
	}
	// base64
	bytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		glog.Errorln("[GetToken]", err)
		return nil, err
	}
	// rsa
	res, err := RSADecrypt(bytes, prikey)
	if err != nil {
		glog.Errorln("[RSADecrypt]", err)
		return nil, err
	}
	var info RoomTokenInfo
	err = json.Unmarshal(res, &info)
	if err != nil {
		glog.Errorln("[json.Unmarshal]", err)
		return nil, err
	}
	return &info, nil
}

// 解密，origData：原始数据，publicKey：公钥
func RSAEncrypt(origData []byte, publicKey []byte) ([]byte, error) {
	// 解密pem格式的公钥
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return nil, errors.New("public key error")
	}
	// 解析公钥
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	// 类型断言
	pub := pubInterface.(*rsa.PublicKey)
	// 加密
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

// 解密，ciphertext：密文，privateKey：私钥
func RSADecrypt(ciphertext []byte, privateKey []byte) ([]byte, error) {
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error")
	}
	// 解析PKCS1格式的私钥
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv, ciphertext)
}
