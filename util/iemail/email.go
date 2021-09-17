package iemail

import (
	"github.com/matcornic/hermes/v2"
	"gopkg.in/gomail.v2"
	"net/mail"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string

	Product Product
}
type Product struct {
	Name        string
	Link        string
	Logo        string
	Copyright   string
	TroubleText string
}
type Header struct {
	From    mail.Address
	To      string
	Subject string
}

type EmailSend struct {
	c *Config
}

func New(c *Config) *EmailSend {
	e := &EmailSend{c: c}
	return e
}

func (e *EmailSend) SendHTML(header Header, content hermes.Email) error {
	typ := "text/html"
	body, err := e.GenBody(typ, content)
	if err != nil {
		return err
	}
	return e.send(header, typ, body)
}

func (e *EmailSend) SendPlainText(header Header, content hermes.Email) error {
	typ := "text/plain"
	body, err := e.GenBody(typ, content)
	if err != nil {
		return err
	}
	return e.send(header, typ, body)
}

func (e *EmailSend) GenBody(typ string, content hermes.Email) (string, error) {
	h := hermes.Hermes{
		Theme:         new(hermes.Default),
		TextDirection: "",
		Product: hermes.Product{
			Name:        e.c.Product.Name,
			Link:        e.c.Product.Link,
			Logo:        e.c.Product.Logo,
			Copyright:   e.c.Product.Copyright,
			TroubleText: e.c.Product.TroubleText,
		},
		DisableCSSInlining: false,
	}
	switch typ {
	case "text/html":
		return h.GenerateHTML(content)
	default:
		return h.GeneratePlainText(content)
	}
}

func (e *EmailSend) send(header Header, typ, body string) error {
	msg := gomail.NewMessage()
	msg.SetHeader("From", header.From.String())
	msg.SetHeader("To", header.To)
	msg.SetHeader("Subject", header.Subject)

	if typ != "" {
		msg.SetBody(typ, body)
	}

	d := gomail.NewDialer(e.c.Host, e.c.Port, e.c.Username, e.c.Password)
	return d.DialAndSend(msg)
}
