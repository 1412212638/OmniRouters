package common

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

var resendEmailClient = &http.Client{
	Timeout: 20 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func ValidateResendConfiguration(apiKey, from string) error {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" || strings.ContainsAny(apiKey, "\r\n\x00") {
		return fmt.Errorf("Resend API key is not configured or invalid")
	}
	if strings.ContainsAny(from, "\r\n\x00") {
		return fmt.Errorf("Resend sender address is not configured or invalid")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(from)); err != nil {
		return fmt.Errorf("Resend sender address is not configured or invalid")
	}
	return nil
}

func sendResendEmail(apiKey, from, subject, receiver, content string) error {
	if err := ValidateResendConfiguration(apiKey, from); err != nil {
		return err
	}
	sender, _ := mail.ParseAddress(strings.TrimSpace(from))
	if sender.Name == "" {
		sender.Name = SystemName
	}
	to := make([]string, 0)
	for _, address := range strings.Split(receiver, ";") {
		recipient, err := mail.ParseAddress(strings.TrimSpace(address))
		if err != nil || len(to) >= 50 {
			return fmt.Errorf("Invalid email address")
		}
		to = append(to, recipient.Address)
	}
	payload, err := Marshal(struct {
		From string `json:"from"`
		To []string `json:"to"`
		Subject string `json:"subject"`
		HTML string `json:"html"`
	}{sender.String(), to, subject, content})
	if err != nil {
		return fmt.Errorf("Resend request failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("Resend request failed")
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(apiKey))
	req.Header.Set("Content-Type", "application/json")
	resp, err := resendEmailClient.Do(req)
	if err != nil {
		SysError(fmt.Sprintf("Resend request failed: error=%q", err))
		return fmt.Errorf("Resend request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var failure struct {
			Name string `json:"name"`
			Message string `json:"message"`
		}
		_ = DecodeJson(io.LimitReader(resp.Body, 4096), &failure)
		SysError(fmt.Sprintf("Resend rejected the email: status=%d name=%q message=%q", resp.StatusCode, failure.Name, failure.Message))
		return fmt.Errorf("Resend email delivery failed (HTTP %d)", resp.StatusCode)
	}
	var result struct { ID string `json:"id"` }
	if err := DecodeJson(io.LimitReader(resp.Body, 4096), &result); err != nil || strings.TrimSpace(result.ID) == "" {
		return fmt.Errorf("Invalid Resend delivery response")
	}
	return nil
}
