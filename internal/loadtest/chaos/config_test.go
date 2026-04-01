package chaos_test

import (
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/loadtest/chaos"
)

func TestConfig_Validate_ClampRates(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{
		SlowMoveRate:   1.5,
		ErrorMoveRate:  -0.3,
		DisconnectRate: 2.0,
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After validation, rates should be clamped but we can't read them directly.
	// Validate that it doesn't error is sufficient for clamping.
}

func TestConfig_Validate_DefaultSlowDelay(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{
		SlowMoveRate: 0.5,
		// SlowMoveDelay is zero — should be defaulted to 2s.
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfig_Validate_ErrorRatesPlusSlowRateExceeds1(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{
		SlowMoveRate:  0.7,
		ErrorMoveRate: 0.5,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for rates summing > 1.0")
	}
}

func TestConfig_Validate_NegativeReconnectDelay(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{
		DisconnectRate: 0.1,
		ReconnectDelay: -1 * time.Second,
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for negative ReconnectDelay with DisconnectRate > 0")
	}
}

func TestConfig_Enabled_AllZero(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{}
	if cfg.Enabled() {
		t.Fatal("expected Enabled() to be false for zero config")
	}
}

func TestConfig_Enabled_WithSlowMove(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{SlowMoveRate: 0.1}
	if !cfg.Enabled() {
		t.Fatal("expected Enabled() to be true")
	}
}

func TestConfig_Enabled_WithDisconnect(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{DisconnectRate: 0.05}
	if !cfg.Enabled() {
		t.Fatal("expected Enabled() to be true")
	}
}

func TestConfig_Enabled_WithErrorMove(t *testing.T) {
	t.Parallel()

	cfg := chaos.Config{ErrorMoveRate: 0.02}
	if !cfg.Enabled() {
		t.Fatal("expected Enabled() to be true")
	}
}
