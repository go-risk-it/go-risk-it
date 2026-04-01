// Package annotations provides a Grafana HTTP API client for posting
// test-phase annotations to dashboards.
package annotations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"go.opentelemetry.io/otel/attribute"
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
	ctx := context.Background()

	if a.grafanaURL == "" {
		return
	}

	body, err := json.Marshal(map[string]any{
		"time": time.Now().UnixMilli(),
		"text": text,
		"tags": tags,
	})
	if err != nil {
		observe.Error(ctx, err, "annotation marshal failed")

		return
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.grafanaURL+"/api/annotations",
		bytes.NewReader(body),
	)
	if err != nil {
		observe.Error(ctx, err, "annotation request creation failed")

		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		observe.Error(ctx, err, "annotation post failed")

		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		observe.Warn(ctx, "annotation unexpected status",
			attribute.Int("status_code", resp.StatusCode),
		)
	}
}
