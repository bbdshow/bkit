package ginutil

import (
	"github.com/bbdshow/bkit/auth/jwt"
	"github.com/gin-gonic/gin"
	"time"
)

var JWTDataKey = "jwt_claims_data"

func GenJWTToken(data string, ttl time.Duration, signingKey ...string) (token string, err error) {
	return jwt.GenerateJWTToken(jwt.NewCustomClaims(data, ttl, signingKey...))
}

func SetJWTDataToContext(c *gin.Context, token string, signingKey ...string) error {
	data, err := jwt.GetCustomData(token, signingKey...)
	if err != nil {
		return err
	}
	c.Set(JWTDataKey, data)
	return nil
}

func GetJWTDataFromContext(c *gin.Context) (string, bool) {
	val, ok := c.Get(JWTDataKey)
	if !ok {
		return "", ok
	}
	data, ok := val.(string)
	if !ok {
		return "", ok
	}
	return data, true
}
