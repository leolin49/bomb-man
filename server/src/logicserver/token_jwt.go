package main

// import (
// 	"time"

// 	"github.com/dgrijalva/jwt-go"
// 	"github.com/golang/glog"
// )

// // 指定加密密钥
// const signingKey = "Aq12wSdE3"

// type Claims struct {
// 	username string
// 	password string
// 	jwt.StandardClaims
// }

// // 根据用户的用户名和密码生成token
// func JWTcreateToken(name, pwd string) (string, error) {
// 	// 设置token的有效时间
// 	nowTime := time.Now()
// 	expireTime := nowTime.Add(5 * time.Minute)

// 	claims := Claims{
// 		username: name,
// 		password: pwd,
// 		StandardClaims: jwt.StandardClaims{
// 			// 过期时间
// 			ExpiresAt: expireTime.Unix(),
// 			// token发行人
// 			Issuer: "LogicServer",
// 		},
// 	}
// 	// 签名方法
// 	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	// 生成完整，已签名的token
// 	token, err := tokenClaims.SignedString([]byte(signingKey))
// 	glog.Errorln("[Token] create token success:", token)
// 	return token, err
// }

// func JWTparseToken(token string) (*Claims, error) {
// 	// 用于解析鉴权的声明，方法内部主要是具体的解码和校验的过程，最终返回*Token
// 	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(t *jwt.Token) (interface{}, error) {
// 		return []byte(signingKey), nil
// 	})
// 	if err != nil {
// 		glog.Infoln("[ParseToken] parse error")
// 		return nil, err
// 	}
// 	if tokenClaims != nil {
// 		// 从tokenClaims中获取到Claims对象，并使用断言，将该对象转换为我们自己定义的Claims
// 		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
// 			return claims, nil
// 		}
// 	}
// 	return nil, err
// }
