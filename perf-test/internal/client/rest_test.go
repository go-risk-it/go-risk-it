package client

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
)

func newTestREST(baseURL string, collector *metrics.Collector) *REST {
	return NewREST(baseURL, "test-token", nil, collector, DefaultRetryConfig())
}

func TestDo_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	r := newTestREST(srv.URL, nil)

	resp, err := r.do(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestDo_RetryOnTransientHTTP(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("unavailable"))

			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("success"))
	}))
	defer srv.Close()

	r := newTestREST(srv.URL, nil)

	resp, err := r.do(http.MethodGet, "/retry", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if got := calls.Load(); got != 3 {
		t.Fatalf("expected 3 calls, got %d", got)
	}
}

func TestDo_RetryExhaustion(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("always unavailable"))
	}))
	defer srv.Close()

	r := newTestREST(srv.URL, nil)

	resp, err := r.do(http.MethodGet, "/exhaust", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	// On the last attempt (attempt == maxRetries-1), the retry check is skipped
	// and the response is returned as-is.
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", resp.StatusCode)
	}

	if got := calls.Load(); got != int64(DefaultRetryConfig().MaxRetries) {
		t.Fatalf("expected %d calls, got %d", DefaultRetryConfig().MaxRetries, got)
	}
}

func TestDo_BodyReplay(t *testing.T) {
	t.Parallel()

	type payload struct {
		Name string `json:"name"`
	}

	var calls atomic.Int64

	var bodiesReceived [2]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := int(calls.Add(1))
		body, _ := io.ReadAll(r.Body)

		if n <= 2 {
			bodiesReceived[n-1] = string(body)
		}

		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	r := newTestREST(srv.URL, nil)

	resp, err := r.do(http.MethodPost, "/body", payload{Name: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	expected, _ := json.Marshal(payload{Name: "test"})

	for i, got := range bodiesReceived {
		if got != string(expected) {
			t.Fatalf("attempt %d: expected body %q, got %q", i+1, expected, got)
		}
	}
}

func TestDo_ConflictNotRetried(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte("conflict"))
	}))
	defer srv.Close()

	r := newTestREST(srv.URL, nil)

	resp, err := r.do(http.MethodPost, "/conflict", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDo_NetworkErrorRetry(t *testing.T) {
	t.Parallel()

	// Create a listener that immediately closes connections.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	go func() {
		for {
			conn, acceptErr := ln.Accept()
			if acceptErr != nil {
				return
			}

			conn.Close()
		}
	}()

	defer ln.Close()

	r := newTestREST("http://"+ln.Addr().String(), nil)

	_, err = r.do(http.MethodGet, "/netfail", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDo_CollectorRecordsRetries(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	collector := metrics.NewCollector(1 * time.Minute)
	r := newTestREST(srv.URL, collector)

	resp, err := r.do(http.MethodGet, "/retries", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	snap := collector.Snapshot()
	if snap.TotalRetries != 2 {
		t.Fatalf("expected 2 retries, got %d", snap.TotalRetries)
	}
}

func TestDo_CollectorRecordsHTTPStatus(t *testing.T) {
	t.Parallel()

	var calls atomic.Int64

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := calls.Add(1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	collector := metrics.NewCollector(1 * time.Minute)
	r := newTestREST(srv.URL, collector)

	resp, err := r.do(http.MethodGet, "/status", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	snap := collector.Snapshot()

	if snap.HTTPStatusCounts[503] != 1 {
		t.Fatalf("expected 1 count for 503, got %d", snap.HTTPStatusCounts[503])
	}

	if snap.HTTPStatusCounts[200] != 1 {
		t.Fatalf("expected 1 count for 200, got %d", snap.HTTPStatusCounts[200])
	}
}
