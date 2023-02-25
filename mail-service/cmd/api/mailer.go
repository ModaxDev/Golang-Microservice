package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"time"
)

type Mail struct {
	Domain      string `json:"domain"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Encryption  string `json:"encryption"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name"`
}

type Message struct {
	From        string   `json:"from"`
	FromName    string   `json:"from_name"`
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Attachments []string `json:"attachments"`
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"message": msg.Data,
	}

	msg.DataMap = data

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To)
	email.SetSubject(msg.Subject)
	email.SetBody(mail.TextHTML, plainMessage)
	email.AddAlternative(mail.TextPlain, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	err = email.Send(smtClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.html.gohtml"

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCss(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) inlineCss(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil

}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	err = t.ExecuteTemplate(&tpl, "body", msg.DataMap)
	if err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) getEncryption(encryption string) mail.Encryption {
	switch encryption {
	case "ssl":
		return mail.EncryptionSSLTLS
	case "tls":
		return mail.EncryptionSTARTTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
