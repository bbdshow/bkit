package ginutil

import (
	"github.com/bbdshow/bkit/auth/jwt"
	"github.com/gin-gonic/gin"
	"time"
)

var JWTDataKey = "jwt_claims_data"

func GenJWTToken(data interface{}, ttl time.Duration, signingKey ...string) (token string, err error) {
	return jwt.GenerateJWTToken(jwt.NewCustomClaims(data, ttl, signingKey...))
}

func SetJWTDataToContext(c *gin.Context, token string, signingKey ...string) error {
	var data interface{}
	if err := jwt.GetCustomData(token, &data, signingKey...); err != nil {
		return err
	}
	c.Set(JWTDataKey, data)
	return nil
}

func GetJWTDataFromContext(c *gin.Context) (data interface{}, exists bool) {
	val, ok := c.Get(JWTDataKey)
	if !ok {
		return data, ok
	}
	return val, true
}
