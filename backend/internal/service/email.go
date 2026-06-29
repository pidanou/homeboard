package service

import (
	"bytes"
	"crypto/tls"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net"
	"net/smtp"
)

//go:embed templates/email.html
var emailTemplate string

var tmpl = template.Must(template.New("email").Parse(emailTemplate))

type emailData struct {
	Subject    string
	Heading    string
	Name       string
	Body       string
	ButtonText string
	ButtonURL  string
}

type EmailService struct {
	host string
	port string
	auth smtp.Auth
	from string
	tls  bool
}

// NewEmailService returns nil if host is empty (email disabled).
func NewEmailService(host, port, user, pass, from string, useTLS bool) *EmailService {
	if host == "" {
		return nil
	}
	if port == "" {
		port = "587"
	}
	if from == "" {
		from = user
	}
	var auth smtp.Auth
	if user != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return &EmailService{host: host, port: port, auth: auth, from: from, tls: useTLS}
}

// Send sends a multipart/alternative email (plain text + HTML). No-ops if e is nil.
func (e *EmailService) Send(to string, data emailData) {
	if e == nil {
		return
	}

	var htmlBuf bytes.Buffer
	if err := tmpl.Execute(&htmlBuf, data); err != nil {
		slog.Error("email template render failed", "err", err)
		return
	}

	plain := fmt.Sprintf("Hi %s,\n\n%s\n", data.Name, data.Body)
	if data.ButtonURL != "" {
		plain += fmt.Sprintf("\n%s: %s\n", data.ButtonText, data.ButtonURL)
	}

	boundary := "==familyboard-email-boundary=="
	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n--%s--",
		e.from, to, data.Subject, boundary, boundary, plain, boundary, htmlBuf.String(), boundary,
	)

	addr := net.JoinHostPort(e.host, e.port)
	var err error
	if e.tls {
		err = e.sendTLS(addr, to, msg)
	} else {
		err = smtp.SendMail(addr, e.auth, e.from, []string{to}, []byte(msg))
	}
	if err != nil {
		slog.Error("email send failed", "to", to, "subject", data.Subject, "err", err)
	}
}

func (e *EmailService) sendTLS(addr, to, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.host})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()
	c, err := smtp.NewClient(conn, e.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Quit()
	if e.auth != nil {
		if err := c.Auth(e.auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := c.Mail(e.from); err != nil {
		return err
	}
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err = fmt.Fprint(w, msg); err != nil {
		return err
	}
	return w.Close()
}
