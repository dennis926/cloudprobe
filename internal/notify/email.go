package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailChannel SMTP邮件通知渠道
type EmailChannel struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	From         string
	UseTLS       bool
	UseSSL       bool
}

func (e *EmailChannel) Name() string {
	return "email"
}

func (e *EmailChannel) Send(ctx context.Context, title, content string, recipient string) error {
	if recipient == "" {
		return fmt.Errorf("email recipient is empty")
	}

	// 构建邮件内容
	headers := make(map[string]string)
	headers["From"] = e.From
	headers["To"] = recipient
	headers["Subject"] = title
	headers["Content-Type"] = "text/html; charset=UTF-8"

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(content)

	addr := fmt.Sprintf("%s:%d", e.SMTPHost, e.SMTPPort)

	// 发送逻辑
	var err error
	if e.UseSSL {
		err = e.sendSSL(addr, recipient, msg.String())
	} else if e.UseTLS {
		err = e.sendTLS(addr, recipient, msg.String())
	} else {
		auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPHost)
		err = smtp.SendMail(addr, auth, e.From, []string{recipient}, []byte(msg.String()))
	}

	return err
}

func (e *EmailChannel) sendTLS(addr, recipient, msg string) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: e.SMTPHost})
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, e.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	auth := smtp.PlainAuth("", e.SMTPUser, e.SMTPPassword, e.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return err
	}
	if err := client.Mail(e.From); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(msg))
	if err != nil {
		return err
	}
	return w.Close()
}

func (e *EmailChannel) sendSSL(addr, recipient, msg string) error {
	return e.sendTLS(addr, recipient, msg)
}

func (e *EmailChannel) Test(ctx context.Context) error {
	return e.Send(ctx, "CloudProbe 邮件测试", "<p>这是一封测试邮件，如果收到说明配置正确。</p>", e.SMTPUser)
}
