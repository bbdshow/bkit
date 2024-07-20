package iemail

import (
	"github.com/matcornic/hermes/v2"
	"net/mail"
	"testing"
)

var m *EmailSend

func TestMain(t *testing.M) {
	c := &Config{
		Host:     "smtp.yandex.com",
		Port:     465,
		Username: "",
		Password: "",
		Product: Product{
			Name:        "I Email",
			Link:        "link.com",
			Logo:        "",
			Copyright:   "",
			TroubleText: "",
		},
	}

	m = New(c)
	t.Run()
}

var (
	from = ""
	to   = ""
)

func TestEmailSend_SendHTML(t *testing.T) {
	e := hermes.Email{
		Body: hermes.Body{
			Name:   "html",
			Intros: []string{"test html"},
			Dictionary: []hermes.Entry{
				{
					Key:   "",
					Value: "",
				},
			},
			Table:        hermes.Table{},
			Actions:      nil,
			Outros:       nil,
			Greeting:     "",
			Signature:    "",
			Title:        "test html",
			FreeMarkdown: "",
		},
	}

	if err := m.SendHTML(Header{
		From: mail.Address{
			Name:    "",
			Address: from,
		},
		To:      to,
		Subject: "",
	}, e); err != nil {
		t.Fatal(err)
	}
}
