package alert

import (
	"context"
	"fmt"
	"testing"
)

func TestTelegram_Send(t *testing.T) {
	tel := NewTelegram()
	tel.SetHookURL(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%d",
		"", -650364623))
	if err := tel.Send(context.Background(), "Send Test Message"); err != nil {
		t.Fatal(err)
	}
}
