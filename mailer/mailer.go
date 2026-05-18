package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/config"
)

type Sender interface {
	SendVerificationCode(ctx context.Context, to string, scene string, code string, ttl time.Duration) error
}

func NewSender(cfg config.MailConfig) (Sender, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "console":
		return ConsoleSender{}, nil
	case "smtp":
		return NewSMTPSender(cfg.SMTP), nil
	case "aliyun":
		return AliyunSender{}, nil
	default:
		return nil, fmt.Errorf("unsupported mail provider: %s", cfg.Provider)
	}
}

type ConsoleSender struct{}

func (ConsoleSender) SendVerificationCode(_ context.Context, to string, scene string, code string, ttl time.Duration) error {
	log.Printf("[mail:console] to=%s scene=%s code=%s ttl=%s", to, scene, code, ttl)
	return nil
}

type AliyunSender struct{}

func (AliyunSender) SendVerificationCode(context.Context, string, string, string, time.Duration) error {
	return errors.New("aliyun mail sender is not implemented yet")
}

type SMTPSender struct {
	cfg config.SMTPConfig
}

func NewSMTPSender(cfg config.SMTPConfig) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) SendVerificationCode(ctx context.Context, to string, scene string, code string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.cfg.Host == "" || s.cfg.Port == 0 || s.cfg.From == "" {
		return errors.New("smtp host, port and from are required")
	}

	message := s.buildMessage(to, scene, code, ttl)
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)

	if s.cfg.UseTLS {
		return s.sendWithTLS(addr, to, message)
	}
	if s.cfg.StartTLS {
		return s.sendWithStartTLS(addr, to, message)
	}

	return smtp.SendMail(addr, s.auth(), s.cfg.From, []string{to}, []byte(message))
}

func (s *SMTPSender) buildMessage(to string, scene string, code string, ttl time.Duration) string {
	subject := "Steins Gate 邮箱验证码"
	if scene == "reset_password" {
		subject = "Steins Gate 修改密码验证码"
	}

	fromAddress := mail.Address{
		Name:    s.cfg.FromName,
		Address: s.cfg.From,
	}
	from := fromAddress.String()
	if s.cfg.FromName == "" {
		from = s.cfg.From
	}

	body := fmt.Sprintf(`你的邮箱验证码是：%s

用途：%s
有效期：%d 分钟

如果不是你本人操作，请忽略这封邮件。
`, code, scene, int(ttl.Minutes()))

	headers := map[string]string{
		"From":                      from,
		"To":                        to,
		"Subject":                   mime.QEncoding.Encode("UTF-8", subject),
		"MIME-Version":              "1.0",
		"Content-Type":              `text/plain; charset="UTF-8"`,
		"Content-Transfer-Encoding": "8bit",
	}

	var builder strings.Builder
	for key, value := range headers {
		builder.WriteString(key)
		builder.WriteString(": ")
		builder.WriteString(value)
		builder.WriteString("\r\n")
	}
	builder.WriteString("\r\n")
	builder.WriteString(body)

	return builder.String()
}

func (s *SMTPSender) sendWithTLS(addr string, to string, message string) error {
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, s.tlsConfig())
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return err
	}
	defer client.Close()

	return s.sendWithClient(client, to, message)
}

func (s *SMTPSender) sendWithStartTLS(addr string, to string, message string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.StartTLS(s.tlsConfig()); err != nil {
		return err
	}

	return s.sendWithClient(client, to, message)
}

func (s *SMTPSender) sendWithClient(client *smtp.Client, to string, message string) error {
	if s.cfg.Username != "" {
		if err := client.Auth(s.auth()); err != nil {
			return err
		}
	}
	if err := client.Mail(s.cfg.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	return client.Quit()
}

func (s *SMTPSender) auth() smtp.Auth {
	if s.cfg.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
}

func (s *SMTPSender) tlsConfig() *tls.Config {
	return &tls.Config{
		ServerName:         s.cfg.Host,
		InsecureSkipVerify: s.cfg.SkipVerify,
	}
}
