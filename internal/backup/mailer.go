package backup

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/smtp"
	"os"
	"path/filepath"
	"time"

	"cloudprobe/internal/config"
)

// Mailer 备份邮件发送器
type Mailer struct {
	smtpHost     string
	smtpPort     int
	smtpUser     string
	smtpPassword string
	from         string
}

// NewMailerFromConfig 从配置创建邮件发送器
func NewMailerFromConfig() *Mailer {
	cfg := config.Get()
	return &Mailer{
		smtpHost:     cfg.SMTP.Host,
		smtpPort:     cfg.SMTP.Port,
		smtpUser:     cfg.SMTP.User,
		smtpPassword: cfg.SMTP.Password,
		from:         cfg.SMTP.From,
	}
}

// SendBackupEmail 发送备份邮件
func (m *Mailer) SendBackupEmail(to string, files []string) error {
	if m.smtpHost == "" || m.smtpUser == "" {
		return fmt.Errorf("SMTP not configured")
	}

	// 构建 multipart 邮件
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 邮件头
	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: CloudProbe 数据备份 - %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=%s\r\n\r\n",
		m.from, to, time.Now().Format("2006-01-02"), writer.Boundary())
	buf.WriteString(headers)

	// 邮件正文
	part, _ := writer.CreatePart(map[string][]string{
		"Content-Type": {"text/plain; charset=UTF-8"},
	})
	body := fmt.Sprintf("CloudProbe 数据备份\n备份时间: %s\n附件数量: %d\n\n本邮件由 CloudProbe 自动发送。",
		time.Now().Format("2006-01-02 15:04:05"), len(files))
	part.Write([]byte(body))

	// 附件
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		filename := filepath.Base(file)
		part, _ := writer.CreateFormFile("attachment", filename)
		part.Write(data)
	}

	writer.Close()

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", m.smtpHost, m.smtpPort)
	auth := smtp.PlainAuth("", m.smtpUser, m.smtpPassword, m.smtpHost)
	return smtp.SendMail(addr, auth, m.from, []string{to}, buf.Bytes())
}
