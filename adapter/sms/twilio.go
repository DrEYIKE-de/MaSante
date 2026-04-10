package sms

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Twilio implements domain.SMSProvider for the Twilio API.
type Twilio struct {
	accountSID string
	authToken  string
	fromNumber string
	client     *http.Client
}

// NewTwilio returns a configured Twilio provider.
func NewTwilio(accountSID, authToken, fromNumber string) *Twilio {
	return &Twilio{
		accountSID: accountSID,
		authToken:  authToken,
		fromNumber: fromNumber,
		client:     &http.Client{},
	}
}

// Send sends an SMS via Twilio.
func (p *Twilio) Send(ctx context.Context, to, message string) (string, error) {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", p.accountSID)

	data := url.Values{
		"To":   {to},
		"From": {p.fromNumber},
		"Body": {message},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.SetBasicAuth(p.accountSID, p.authToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("twilio: status %d: %s", resp.StatusCode, body)
	}

	return string(body), nil
}

// Name returns the provider name.
func (p *Twilio) Name() string { return "twilio" }
