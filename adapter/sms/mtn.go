package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MTN implements domain.SMSProvider for the MTN SMS API.
type MTN struct {
	apiKey   string
	secret   string
	senderID string
	baseURL  string
	client   *http.Client
}

// NewMTN returns a configured MTN SMS provider.
func NewMTN(apiKey, secret, senderID string) *MTN {
	return &MTN{
		apiKey:   apiKey,
		secret:   secret,
		senderID: senderID,
		baseURL:  "https://api.mtn.com/v1/sms/outbound",
		client:   &http.Client{},
	}
}

// Send sends an SMS via MTN.
func (p *MTN) Send(ctx context.Context, to, message string) (string, error) {
	payload := map[string]any{
		"senderAddress": p.senderID,
		"receiverAddress": []string{to},
		"message":         message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("mtn: status %d: %s", resp.StatusCode, respBody)
	}

	return string(respBody), nil
}

// Name returns the provider name.
func (p *MTN) Name() string { return "mtn" }
