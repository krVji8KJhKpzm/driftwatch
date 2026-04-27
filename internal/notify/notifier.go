// Package notify provides webhook and notification dispatch for drift events.
package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/driftwatch/internal/drift"
)

// Channel represents a supported notification channel type.
type Channel string

const (
	ChannelWebhook Channel = "webhook"
	ChannelSlack   Channel = "slack"
)

// Config holds the notification configuration.
type Config struct {
	Channel  Channel `json:"channel"`
	URL      string  `json:"url"`
	OnDrift  bool    `json:"on_drift"`
	OnClean  bool    `json:"on_clean"`
	Timeout  int     `json:"timeout_seconds"`
}

// Payload is the JSON body sent to the webhook endpoint.
type Payload struct {
	Timestamp  time.Time        `json:"timestamp"`
	DriftCount int              `json:"drift_count"`
	CleanCount int              `json:"clean_count"`
	Containers []ContainerEvent `json:"containers"`
}

// ContainerEvent summarises drift state for a single container.
type ContainerEvent struct {
	Name    string `json:"name"`
	Drifted bool   `json:"drifted"`
	Fields  []string `json:"drifted_fields,omitempty"`
}

// Notifier dispatches notifications based on drift results.
type Notifier struct {
	cfg    *Config
	client *http.Client
}

// New returns a Notifier configured from cfg.
func New(cfg *Config) *Notifier {
	timeout := 10
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	return &Notifier{
		cfg: cfg,
		client: &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

// Send evaluates results against the config and dispatches a notification
// when the conditions (on_drift / on_clean) are met.
func (n *Notifier) Send(results []drift.Result) error {
	if n.cfg == nil || n.cfg.URL == "" {
		return nil
	}

	payload := buildPayload(results)

	hasDrift := payload.DriftCount > 0
	if hasDrift && !n.cfg.OnDrift {
		return nil
	}
	if !hasDrift && !n.cfg.OnClean {
		return nil
	}

	switch n.cfg.Channel {
	case ChannelSlack:
		return n.sendSlack(payload)
	default:
		return n.sendWebhook(payload)
	}
}

// sendWebhook posts the payload as JSON to the configured URL.
func (n *Notifier) sendWebhook(payload Payload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("notify: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.cfg.URL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("notify: post webhook: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: webhook returned status %d", resp.StatusCode)
	}
	return nil
}

// sendSlack formats a simple Slack-compatible message and posts it.
func (n *Notifier) sendSlack(payload Payload) error {
	text := fmt.Sprintf(":mag: *DriftWatch* — %d drifted, %d clean containers at %s",
		payload.DriftCount, payload.CleanCount, payload.Timestamp.Format(time.RFC3339))

	slackBody, err := json.Marshal(map[string]string{"text": text})
	if err != nil {
		return fmt.Errorf("notify: marshal slack payload: %w", err)
	}

	resp, err := n.client.Post(n.cfg.URL, "application/json", bytes.NewReader(slackBody))
	if err != nil {
		return fmt.Errorf("notify: post slack: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notify: slack returned status %d", resp.StatusCode)
	}
	return nil
}

// buildPayload converts drift results into a Payload.
func buildPayload(results []drift.Result) Payload {
	p := Payload{Timestamp: time.Now().UTC()}
	for _, r := range results {
		fields := make([]string, 0, len(r.Diffs))
		for _, d := range r.Diffs {
			fields = append(fields, d.Field)
		}
		if r.Drifted {
			p.DriftCount++
		} else {
			p.CleanCount++
		}
		p.Containers = append(p.Containers, ContainerEvent{
			Name:    r.Name,
			Drifted: r.Drifted,
			Fields:  fields,
		})
	}
	return p
}
