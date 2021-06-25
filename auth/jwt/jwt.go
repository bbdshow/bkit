package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	Issuer                          = "Anonymous"
	SigningKey                      = []byte("Please replace the default signing key")
	SigningMethod jwt.SigningMethod = jwt.SigningMethodHS256

	ErrCustomClaimsInValid = errors.New("custom claims invalid")
)

func SetIssuer(issuer string) {
	Issuer = issuer
}

func SetSigningMethod(method jwt.SigningMethod) {
	SigningMethod = method
}

type CustomClaims struct {
	SigningKey []byte
	Data       string `json:"data"` // 任意字符串
	jwt.StandardClaims
}

// 自定义携带数据结构
type CustomData string

// NewCustomClaims 自定义请求权
func NewCustomClaims(data string, ttl time.Duration, signingKey ...string) *CustomClaims {
	cc := &CustomClaims{
		SigningKey: SigningKey,
		Data:       data,
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Add(-1 * time.Second).Unix(),
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Issuer:    Issuer,
		},
	}
	if len(signingKey) > 0 {
		cc.SigningKey = []byte(signingKey[0])
	}
	return cc
}

func GenerateJWTToken(customClaims *CustomClaims) (string, error) {
	token := jwt.NewWithClaims(SigningMethod, customClaims)
	if customClaims.SigningKey == nil {
		customClaims.SigningKey = SigningKey
	}
	str, err := token.SignedString(customClaims.SigningKey)
	if err != nil {
		return "", err
	}
	return str, nil
}

func VerifyJWTToken(tokenStr string, signingKey ...string) (bool, error) {
	key := SigningKey
	if len(signingKey) > 0 {
		key = []byte(signingKey[0])
	}
	token, err := parseJWTToken(tokenStr, key)
	if err != nil {
		return false, err
	}
	return token.Valid, nil
}

func GetCustomData(tokenStr string, signingKey ...string) (string, error) {
	key := SigningKey
	if len(signingKey) > 0 {
		key = []byte(signingKey[0])
	}
	token, err := parseJWTToken(tokenStr, key)
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", ErrCustomClaimsInValid
	}
	return claims.Data, err
}

func parseJWTToken(tokenStr string, signingKey []byte) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != SigningMethod.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
