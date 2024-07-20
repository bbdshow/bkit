package bkit

import (
	"context"
	"fmt"
	"testing"
)

func TestNotify_Telegram(t *testing.T) {
	Notify.SetProxy("http://127.0.0.1:7890")
	hookURL := (fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d",
		"", 0))
	if err := Notify.SendTelegramTextMsg(context.Background(), hookURL, "Send Test Message", true); err != nil {
		t.Fatal(err)
	}
}
