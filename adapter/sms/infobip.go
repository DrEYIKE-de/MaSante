package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Infobip implements domain.SMSProvider for the Infobip API.
type Infobip struct {
	apiKey   string
	baseURL  string
	senderID string
	client   *http.Client
}

// NewInfobip returns a configured Infobip provider.
func NewInfobip(apiKey, baseURL, senderID string) *Infobip {
	if baseURL == "" {
		baseURL = "https://api.infobip.com"
	}
	return &Infobip{
		apiKey:   apiKey,
		baseURL:  baseURL,
		senderID: senderID,
		client:   &http.Client{},
	}
}

// Send sends an SMS via Infobip.
func (p *Infobip) Send(ctx context.Context, to, message string) (string, error) {
	payload := map[string]any{
		"messages": []map[string]any{
			{
				"from": p.senderID,
				"destinations": []map[string]string{
					{"to": to},
				},
				"text": message,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/sms/2/text/advanced", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "App "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("infobip: status %d: %s", resp.StatusCode, respBody)
	}

	return string(respBody), nil
}

// Name returns the provider name.
func (p *Infobip) Name() string { return "infobip" }
