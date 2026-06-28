package smartwatch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

type EventHub struct {
	mu          sync.Mutex
	subscribers map[chan Event]struct{}
	webhookURL  string
	client      *http.Client
}

func NewEventHub(webhookURL string) *EventHub {
	return &EventHub{
		subscribers: map[chan Event]struct{}{},
		webhookURL:  webhookURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *EventHub) Publish(ctx context.Context, event Event) {
	h.mu.Lock()
	for ch := range h.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
	h.mu.Unlock()

	if h.webhookURL != "" {
		h.postWebhook(ctx, event)
	}
}

func (h *EventHub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan Event, 16)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		h.mu.Unlock()
		close(ch)
	}()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-ch:
			data, err := json.Marshal(event)
			if err != nil {
				slog.Debug("failed to encode event", "error", err)
				continue
			}
			fmt.Fprintf(w, "event: %s\n", event.Type)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func (h *EventHub) postWebhook(ctx context.Context, event Event) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	data, err := json.Marshal(event)
	if err != nil {
		slog.Debug("failed to encode webhook event", "error", err)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.webhookURL, bytes.NewReader(data))
	if err != nil {
		slog.Debug("failed to build webhook request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Debug("webhook delivery failed", "error", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.Debug("webhook delivery returned non-2xx", "status", resp.StatusCode)
	}
}
