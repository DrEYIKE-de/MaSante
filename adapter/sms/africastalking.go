// Package sms provides driven adapters implementing domain.SMSProvider
// for various African and global SMS gateway providers.
package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AfricasTalking implements domain.SMSProvider for the Africa's Talking API.
type AfricasTalking struct {
	apiKey   string
	username string
	senderID string
	baseURL  string
	client   *http.Client
}

// NewAfricasTalking returns a configured Africa's Talking provider.
func NewAfricasTalking(apiKey, username, senderID string) *AfricasTalking {
	return &AfricasTalking{
		apiKey:   apiKey,
		username: username,
		senderID: senderID,
		baseURL:  "https://api.africastalking.com/version1/messaging",
		client:   &http.Client{},
	}
}

// Send sends an SMS via Africa's Talking.
func (p *AfricasTalking) Send(ctx context.Context, to, message string) (string, error) {
	data := url.Values{
		"username": {p.username},
		"to":       {to},
		"message":  {message},
	}
	if p.senderID != "" {
		data.Set("from", p.senderID)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("apiKey", p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("africastalking: status %d: %s", resp.StatusCode, body)
	}

	return string(body), nil
}

// Name returns the provider name.
func (p *AfricasTalking) Name() string { return "africastalking" }
