package bkit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

var Notify = &NotifyUtil{}

type NotifyUtil struct {
	proxyClient *http.Client
}

func (nu *NotifyUtil) SetProxy(proxy string) {
	if nu.proxyClient == nil {
		nu.proxyClient = &http.Client{}
	}
	nu.proxyClient.Transport = &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(proxy)
		},
	}
}

// SendDingDingText 发送钉钉文本消息
// content: 消息内容
// hookURL: 钉钉机器人hook地址 eg: https://oapi.dingtalk.com/robot/send?access_token=xxxxx
func (nu *NotifyUtil) SendDingDingTextMsg(ctx context.Context, hookURL, content string) error {
	reqBody := struct {
		MsgType string `json:"msgtype"`
		Text    struct {
			Content string `json:"content"`
		} `json:"text"`
	}{}
	respBody := struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}
	reqBody.MsgType = "text"
	reqBody.Text.Content = content
	byt, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", hookURL, bytes.NewReader(byt))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byt, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(byt, &respBody); err != nil {
		return err
	}

	if respBody.ErrCode != 0 {
		return fmt.Errorf("send dingding msg error %d %s", respBody.ErrCode, respBody.ErrMsg)
	}
	return nil
}

func (nu *NotifyUtil) formatTelegram(hookURL string) (_url string, chatID int64, err error) {
	_url = hookURL
	u, err := url.Parse(hookURL)
	if err == nil {
		cid := u.Query().Get("chat_id")
		chatID, err = strconv.ParseInt(cid, 10, 64)
		if err != nil {
			return "", 0, fmt.Errorf("telegram chatID invalid %s", cid)
		}
		u.ForceQuery = false
		_url = u.String()
	}
	return _url, chatID, err
}

// SendTelegramTextMsg 发送Telegram文本消息
// content: 消息内容
// hookURL: Telegram机器人hook地址 eg:https://api.telegram.org/bot{token}/sendMessage?chat_id={chatId}
func (nu *NotifyUtil) SendTelegramTextMsg(ctx context.Context, hookURL, content string, useProxy ...bool) error {
	_url, chatID, err := nu.formatTelegram(hookURL)
	if err != nil {
		return fmt.Errorf("hookURL invalid %w", err)
	}

	reqBody := struct {
		ChatId int64  `json:"chat_id"`
		Text   string `json:"text"`
	}{}
	respBody := struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		ErrMsg  string `json:"description"`
	}{}
	reqBody.ChatId = chatID
	reqBody.Text = content
	byt, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", _url, bytes.NewReader(byt))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	if len(useProxy) > 0 && useProxy[0] && nu.proxyClient != nil {
		client = nu.proxyClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byt, err = io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(byt, &respBody); err != nil {
		return err
	}

	if !respBody.OK {
		return fmt.Errorf("send telegram msg error %d %s", respBody.ErrCode, respBody.ErrMsg)
	}
	return nil
}
