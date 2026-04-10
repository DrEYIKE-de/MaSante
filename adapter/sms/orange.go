package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Orange implements domain.SMSProvider for the Orange SMS API.
type Orange struct {
	apiKey   string
	senderID string
	baseURL  string
	client   *http.Client
}

// NewOrange returns a configured Orange SMS provider.
func NewOrange(apiKey, senderID string) *Orange {
	return &Orange{
		apiKey:   apiKey,
		senderID: senderID,
		baseURL:  "https://api.orange.com/smsmessaging/v1/outbound",
		client:   &http.Client{},
	}
}

// Send sends an SMS via Orange.
func (p *Orange) Send(ctx context.Context, to, message string) (string, error) {
	endpoint := fmt.Sprintf("%s/tel:%s/requests", p.baseURL, p.senderID)

	payload := map[string]any{
		"outboundSMSMessageRequest": map[string]any{
			"address":        "tel:" + to,
			"senderAddress":  "tel:" + p.senderID,
			"outboundSMSTextMessage": map[string]string{
				"message": message,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
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
		return "", fmt.Errorf("orange: status %d: %s", resp.StatusCode, respBody)
	}

	return string(respBody), nil
}

// Name returns the provider name.
func (p *Orange) Name() string { return "orange" }
