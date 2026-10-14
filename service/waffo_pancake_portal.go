package service

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting"
)

const waffoPancakeSessionTokenPath = "/v1/actions/auth/issue-session-token"

type waffoPancakePortalRequest struct {
	StoreID       string `json:"storeId"`
	BuyerIdentity string `json:"buyerIdentity"`
	Purpose       string `json:"purpose"`
}

type waffoPancakePortalResponse struct {
	Data struct {
		Token    string `json:"token"`
		PortalURL string `json:"portalUrl"`
	} `json:"data"`
}

// CreateWaffoPancakeCustomerPortalLink mints a short-lived portal credential
// immediately before use. The returned URL must never be persisted or logged.
func CreateWaffoPancakeCustomerPortalLink(ctx context.Context, buyerIdentity string) (string, error) {
	if strings.TrimSpace(setting.WaffoPancakeStoreID) == "" || strings.TrimSpace(buyerIdentity) == "" {
		return "", fmt.Errorf("Waffo Pancake portal is not configured")
	}
	key, err := parseWaffoPancakePrivateKey(setting.WaffoPancakePrivateKey)
	if err != nil {
		return "", fmt.Errorf("parse Waffo Pancake private key: %w", err)
	}
	body, err := common.Marshal(waffoPancakePortalRequest{
		StoreID: setting.WaffoPancakeStoreID, BuyerIdentity: buyerIdentity, Purpose: "portal",
	})
	if err != nil {
		return "", fmt.Errorf("marshal Waffo Pancake portal request: %w", err)
	}
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	bodyHash := sha256.Sum256(body)
	canonical := "POST\n" + waffoPancakeSessionTokenPath + "\n" + timestamp + "\n" + base64.StdEncoding.EncodeToString(bodyHash[:])
	canonicalHash := sha256.Sum256([]byte(canonical))
	signature, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, canonicalHash[:])
	if err != nil {
		return "", fmt.Errorf("sign Waffo Pancake portal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.waffo.ai"+waffoPancakeSessionTokenPath, strings.NewReader(string(body)))
	if err != nil {
		return "", fmt.Errorf("build Waffo Pancake portal request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Merchant-Id", setting.WaffoPancakeMerchantID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", base64.StdEncoding.EncodeToString(signature))
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("request Waffo Pancake portal link: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read Waffo Pancake portal response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("Waffo Pancake portal request returned status %d", resp.StatusCode)
	}
	var parsed waffoPancakePortalResponse
	if err := common.Unmarshal(responseBody, &parsed); err != nil {
		return "", fmt.Errorf("decode Waffo Pancake portal response: %w", err)
	}
	if parsed.Data.Token == "" || parsed.Data.PortalURL == "" {
		return "", fmt.Errorf("Waffo Pancake portal response is incomplete")
	}
	portalURL, err := url.Parse(parsed.Data.PortalURL)
	if err != nil || portalURL.Scheme != "https" || portalURL.Host != "pancake.waffo.ai" || portalURL.User != nil {
		return "", fmt.Errorf("Waffo Pancake portal URL is invalid")
	}
	portalURL.Fragment = "token=" + parsed.Data.Token
	return portalURL.String(), nil
}

func parseWaffoPancakePrivateKey(raw string) (*rsa.PrivateKey, error) {
	raw = strings.TrimSpace(strings.ReplaceAll(raw, "\\n", "\n"))
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(raw), ""))
		if err != nil {
			return nil, err
		}
		block = &pem.Block{Type: "PRIVATE KEY", Bytes: decoded}
	}
	if block.Type == "RSA PRIVATE KEY" {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}
	return rsaKey, nil
}
