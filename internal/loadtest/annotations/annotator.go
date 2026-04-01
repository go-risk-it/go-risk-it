// Package annotations provides a Grafana HTTP API client for posting
// test-phase annotations to dashboards.
package annotations

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// Annotator posts annotations to the Grafana HTTP API.
// An empty grafanaURL creates a no-op annotator.
type Annotator struct {
	grafanaURL string
	client     *http.Client
}

// NewAnnotator creates an annotator. Empty grafanaURL disables all calls.
func NewAnnotator(grafanaURL string) *Annotator {
	return &Annotator{
		grafanaURL: grafanaURL,
		client:     &http.Client{Timeout: 5 * time.Second},
	}
}

// Annotate posts a timestamped annotation with the given tags.
// Fire-and-forget: errors are logged, not returned.
func (a *Annotator) Annotate(text string, tags ...string) {
	if a.grafanaURL == "" {
		return
	}

	body, _ := json.Marshal(map[string]any{
		"time": time.Now().UnixMilli(),
		"text": text,
		"tags": tags,
	})

	resp, err := a.client.Post(
		a.grafanaURL+"/api/annotations",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[annotations] failed to post: %v", err)

		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[annotations] unexpected status: %d", resp.StatusCode)
	}
}
