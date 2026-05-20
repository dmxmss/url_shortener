package events

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
)

const redirectSubject = "redirect.events"

type RedirectEvent struct {
	ShortCode  string    `json:"short_code"`
	AccessedAt time.Time `json:"accessed_at"`
	UserAgent  string    `json:"user_agent,omitempty"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
}

type Publisher interface {
	PublishRedirect(ctx context.Context, event RedirectEvent) error
	Close()
}

type noopPublisher struct{}

func NewNoop() Publisher {
	return noopPublisher{}
}

func (noopPublisher) PublishRedirect(context.Context, RedirectEvent) error {
	return nil
}

func (noopPublisher) Close() {}

type NATSPublisher struct {
	conn *nats.Conn
}

func NewNATS(url string) (*NATSPublisher, error) {
	conn, err := nats.Connect(url, nats.Name("redirect-service"))
	if err != nil {
		return nil, err
	}
	return &NATSPublisher{conn: conn}, nil
}

func (p *NATSPublisher) PublishRedirect(ctx context.Context, event RedirectEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- p.conn.Publish(redirectSubject, payload)
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (p *NATSPublisher) Close() {
	p.conn.Close()
}
