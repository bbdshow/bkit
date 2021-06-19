package ginutil

import (
	"github.com/bbdshow/bkit/auth/jwt"
	"github.com/gin-gonic/gin"
	"time"
)

type JWTData struct {
	Uid      int64       `json:"uid"`
	Role     string      `json:"role"`
	NickName string      `json:"nickName"`
	Phone    string      `json:"phone"`
	Data     interface{} `json:"data"`
}

var JWTDataKey = "jwt_data_key"

func GenJWTToken(data JWTData, ttl time.Duration, signingKey ...string) (token string, err error) {
	return jwt.GenerateJWTToken(jwt.NewCustomClaims(data, ttl, signingKey...))
}

func SetJWTDataToContext(c *gin.Context, token string, signingKey ...string) error {
	data := JWTData{}
	if err := jwt.GetCustomData(token, &data, signingKey...); err != nil {
		return err
	}
	c.Set(JWTDataKey, data)
	return nil
}

func GetJWTDataFromContext(c *gin.Context) (data JWTData, exists bool) {
	val, ok := c.Get(JWTDataKey)
	if !ok {
		return data, ok
	}
	data, ok = val.(JWTData)
	if !ok {
		return data, ok
	}
	return data, true
}

func JWTDataUidIsEqual(c *gin.Context, uid int64) bool {
	data, exists := GetJWTDataFromContext(c)
	if !exists {
		return exists
	}
	return data.Uid == uid
}
