package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type Telegram struct {
	hookURL string
	chatId  string
	client  *http.Client
}

func NewTelegram() *Telegram {
	return &Telegram{
		client: &http.Client{},
	}
}

// SetHookURL setting RequestURL eg:https://api.telegram.org/bot{token}/sendMessage?chat_id={chatId}
func (tel *Telegram) SetHookURL(_url string) {
	// decode chatId
	u, err := url.Parse(_url)
	if err == nil {
		chatId := u.Query().Get("chat_id")
		tel.chatId = chatId
		u.ForceQuery = false
		tel.hookURL = u.String()
	} else {
		tel.hookURL = _url
	}
}

// Send text message
func (tel *Telegram) Send(ctx context.Context, content string) error {
	chatId, err := strconv.ParseInt(tel.chatId, 10, 64)
	if err != nil {
		return fmt.Errorf("telegram chatId invalid %s", tel.chatId)
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
	reqBody.ChatId = chatId
	reqBody.Text = content
	byt, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tel.hookURL, bytes.NewReader(byt))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := tel.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	byt, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(byt))
	if err := json.Unmarshal(byt, &respBody); err != nil {
		return err
	}

	if !respBody.OK {
		return fmt.Errorf("%s method error %d %s", tel.Method(), respBody.ErrCode, respBody.ErrMsg)
	}
	return nil
}

// Method alert method
func (tel *Telegram) Method() string {
	return "Telegram"
}
