package bkit

import (
	"time"
)

var Token = NewTokenUtil("random_token_secret")

type TokenUtil struct {
	secret      string
	minValidity time.Duration // token 最小有效期
}

func NewTokenUtil(secret string, minValidity ...time.Duration) *TokenUtil {
	min := time.Minute * 60
	if len(minValidity) > 0 && minValidity[0] > 0 {
		min = minValidity[0]
	}
	return &TokenUtil{
		secret:      secret,
		minValidity: min,
	}
}

// GenRandomToken 生成随机 token , 用于安全校验
// 根据 minValidity 计算当前时间的区间，生成 token，
func (t *TokenUtil) GenRandomToken() string {
	now := time.Now()
	start := now.Truncate(t.minValidity)
	end := start.Add(t.minValidity)

	return Str.MD5(t.secret, start.Format(time.DateTime), end.Format(time.DateTime))
}

// VerifyRandomToken 验证随机 token, 时间区间为 minValidity，防止时间回退和时间跳跃
func (t *TokenUtil) VerifyRandomToken(token string) bool {
	now := time.Now()
	start := now.Truncate(t.minValidity)
	end := start.Add(t.minValidity)

	currentToken := Str.MD5(t.secret, start.Format(time.DateTime), end.Format(time.DateTime))
	if token == currentToken {
		return true
	}
	// 防止时间跳跃
	preStart := start.Add(-t.minValidity)
	preEnd := start
	preToken := Str.MD5(t.secret, preStart.Format(time.DateTime), preEnd.Format(time.DateTime))
	return token == preToken
}
