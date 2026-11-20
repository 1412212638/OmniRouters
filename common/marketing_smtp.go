package common

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

const MarketingSMTPOption = "MarketingSMTPSettingsSecret"

type MarketingSMTPConfig struct {
	Mode        string `json:"mode"`
	Server      string `json:"server"`
	Port        int    `json:"port"`
	Account     string `json:"account"`
	Password    string `json:"password,omitempty"`
	From        string `json:"from"`
	DisplayName string `json:"display_name"`
	ReplyTo     string `json:"reply_to"`
	PerSecond   int    `json:"per_second"`
}

func DecodeMarketingSMTPConfig(raw string) (MarketingSMTPConfig, error) {
	cfg := MarketingSMTPConfig{Mode: "inherit", Port: 465, PerSecond: 5}
	if raw != "" {
		if err := UnmarshalJsonStr(raw, &cfg); err != nil {
			return cfg, errors.New("invalid stored marketing SMTP settings")
		}
	}
	return cfg, nil
}

func (cfg *MarketingSMTPConfig) Validate() error {
	cfg.Server = strings.TrimSpace(cfg.Server)
	cfg.Account = strings.TrimSpace(cfg.Account)
	cfg.From = strings.TrimSpace(cfg.From)
	cfg.ReplyTo = strings.TrimSpace(cfg.ReplyTo)
	cfg.DisplayName = strings.TrimSpace(cfg.DisplayName)
	if cfg.Mode != "inherit" && cfg.Mode != "smtp" {
		return errors.New("invalid marketing email delivery mode")
	}
	if cfg.PerSecond < 1 || cfg.PerSecond > 20 {
		return errors.New("marketing email rate must be between 1 and 20 per second")
	}
	if cfg.Mode == "inherit" {
		return nil
	}
	if cfg.Port < 1 || cfg.Port > 65535 || cfg.Server == "" || strings.ContainsAny(cfg.Server, " /:\\@\r\n\t\x00") {
		return errors.New("invalid marketing SMTP server or port")
	}
	if cfg.From == "" {
		cfg.From = cfg.Account
	}
	if !IsPlainEmailAddress(cfg.Account) || !IsPlainEmailAddress(cfg.From) || cfg.Password == "" ||
		(cfg.ReplyTo != "" && !IsPlainEmailAddress(cfg.ReplyTo)) || strings.ContainsAny(cfg.DisplayName, "\r\n\x00") {
		return errors.New("invalid marketing SMTP credentials or sender addresses")
	}
	return nil
}

func IsPlainEmailAddress(value string) bool {
	addr, err := mail.ParseAddress(value)
	return err == nil && addr.Address == value && !strings.ContainsAny(value, "\r\n\x00")
}

type MarketingSMTPError struct {
	Message   string
	Retryable bool
	Unknown   bool
	Fatal     bool
}

func (err *MarketingSMTPError) Error() string { return err.Message }

func marketingSMTPError(err error, stage string) error {
	var reply *textproto.Error
	if errors.As(err, &reply) {
		retry := reply.Code/100 == 4 || strings.Contains(reply.Msg, "RequestLimitExceeded.SmtpRateLimit")
		return &MarketingSMTPError{
			Message:   fmt.Sprintf("marketing SMTP %s rejected (SMTP %d)", stage, reply.Code),
			Retryable: retry,
			Fatal:     !retry && (stage == "login" || stage == "sender"),
		}
	}
	if stage == "submit" {
		return &MarketingSMTPError{Message: "marketing SMTP did not confirm submission", Unknown: true}
	}
	var networkError net.Error
	retry := errors.As(err, &networkError) && networkError.Timeout()
	return &MarketingSMTPError{
		Message:   "marketing SMTP " + stage + " failed",
		Retryable: retry,
		Fatal:     !retry && (stage == "connect" || stage == "login" || stage == "sender"),
	}
}

func (cfg MarketingSMTPConfig) Send(ctx context.Context, subject, receiver, content, unsubscribeURL, messageID, callbackProof string) error {
	if !IsPlainEmailAddress(receiver) {
		return errors.New("invalid marketing email recipient")
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	addr := net.JoinHostPort(cfg.Server, strconv.Itoa(cfg.Port))
	dialer := tls.Dialer{Config: &tls.Config{ServerName: cfg.Server, MinVersion: tls.VersionTLS12}}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return marketingSMTPError(err, "connect")
	}
	defer conn.Close()
	stop := context.AfterFunc(ctx, func() { _ = conn.Close() })
	defer stop()
	if deadline, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(deadline)
	}
	client, err := smtp.NewClient(conn, cfg.Server)
	if err != nil {
		return marketingSMTPError(err, "connect")
	}
	defer client.Close()
	auth := &smtpAutoAuth{username: cfg.Account, password: cfg.Password, host: cfg.Server}
	if err := client.Auth(auth); err != nil {
		return marketingSMTPError(err, "login")
	}
	if err := client.Mail(cfg.From); err != nil {
		return marketingSMTPError(err, "sender")
	}
	if err := client.Rcpt(receiver); err != nil {
		return marketingSMTPError(err, "recipient")
	}
	data, err := client.Data()
	if err != nil {
		return marketingSMTPError(err, "body")
	}
	if err := cfg.writeMessage(data, subject, receiver, content, unsubscribeURL, messageID, callbackProof); err != nil {
		return marketingSMTPError(err, "body")
	}
	if err := data.Close(); err != nil {
		return marketingSMTPError(err, "submit")
	}
	return nil
}

func (cfg MarketingSMTPConfig) writeMessage(data io.Writer, subject, receiver, content, unsubscribeURL, messageID, callbackProof string) error {
	name := cfg.DisplayName
	if name == "" {
		name = SystemName
	}
	sender := (&mail.Address{Name: name, Address: cfg.From}).String()
	encodedSubject := strings.ReplaceAll(strings.ReplaceAll(subject, "\r", " "), "\n", " ")
	encodedSubject = mime.QEncoding.Encode("UTF-8", encodedSubject)
	var headers strings.Builder
	fmt.Fprintf(&headers, "From: %s\r\nTo: <%s>\r\nSubject: %s\r\nDate: %s\r\nMessage-ID: <%s>\r\n", sender, receiver, encodedSubject, time.Now().Format(time.RFC1123Z), messageID)
	if cfg.ReplyTo != "" {
		fmt.Fprintf(&headers, "Reply-To: <%s>\r\n", cfg.ReplyTo)
	}
	if strings.HasPrefix(unsubscribeURL, "https://") {
		fmt.Fprintf(&headers, "List-Unsubscribe: <%s>\r\nList-Unsubscribe-Post: List-Unsubscribe=One-Click\r\n", unsubscribeURL)
	}
	if callbackProof != "" {
		fmt.Fprintf(&headers, "X-Tencentcloudses-Cb-Proof: %s\r\n", callbackProof)
	}
	headers.WriteString("MIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n")
	if _, err := data.Write([]byte(headers.String())); err != nil {
		return err
	}
	body := quotedprintable.NewWriter(data)
	if _, err := body.Write([]byte(content)); err != nil {
		return err
	}
	return body.Close()
}
