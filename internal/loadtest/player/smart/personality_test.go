package smart //nolint:testpackage // whitebox tests for unexported functions

import (
	"testing"
	"time"
)

func TestPersonalities(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		build             func() Personality
		wantName          string
		wantThinkMin      time.Duration
		wantThinkMax      time.Duration
		wantAggression    float64
		wantBaseRatio     float64
		wantContWeight    float64
		wantLargeDiscount float64
		wantContinueAfter float64
	}{
		{
			name:              "Beginner",
			build:             Beginner,
			wantName:          "beginner",
			wantThinkMin:      500 * time.Millisecond,
			wantThinkMax:      2 * time.Second,
			wantAggression:    0.3,
			wantBaseRatio:     1.5,
			wantContWeight:    1.0,
			wantLargeDiscount: 1.0,
			wantContinueAfter: 0.0,
		},
		{
			name:              "Normal",
			build:             Normal,
			wantName:          "normal",
			wantThinkMin:      200 * time.Millisecond,
			wantThinkMax:      800 * time.Millisecond,
			wantAggression:    0.7,
			wantBaseRatio:     1.3,
			wantContWeight:    1.5,
			wantLargeDiscount: 0.9,
			wantContinueAfter: 0.5,
		},
		{
			name:              "Expert",
			build:             Expert,
			wantName:          "expert",
			wantThinkMin:      50 * time.Millisecond,
			wantThinkMax:      200 * time.Millisecond,
			wantAggression:    0.9,
			wantBaseRatio:     1.0,
			wantContWeight:    2.0,
			wantLargeDiscount: 0.8,
			wantContinueAfter: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			p := tt.build()

			if p.Name != tt.wantName {
				t.Fatalf("Name = %q, want %q", p.Name, tt.wantName)
			}

			if p.ThinkTimeMin != tt.wantThinkMin {
				t.Fatalf("ThinkTimeMin = %v, want %v", p.ThinkTimeMin, tt.wantThinkMin)
			}

			if p.ThinkTimeMax != tt.wantThinkMax {
				t.Fatalf("ThinkTimeMax = %v, want %v", p.ThinkTimeMax, tt.wantThinkMax)
			}

			if p.Aggression != tt.wantAggression {
				t.Fatalf("Aggression = %v, want %v", p.Aggression, tt.wantAggression)
			}

			if p.BaseAttackRatio != tt.wantBaseRatio {
				t.Fatalf("BaseAttackRatio = %v, want %v", p.BaseAttackRatio, tt.wantBaseRatio)
			}

			if p.ContinentWeight != tt.wantContWeight {
				t.Fatalf("ContinentWeight = %v, want %v", p.ContinentWeight, tt.wantContWeight)
			}

			if p.LargeArmyDiscount != tt.wantLargeDiscount {
				t.Fatalf(
					"LargeArmyDiscount = %v, want %v",
					p.LargeArmyDiscount,
					tt.wantLargeDiscount,
				)
			}

			if p.ContinueAfterCard != tt.wantContinueAfter {
				t.Fatalf(
					"ContinueAfterCard = %v, want %v",
					p.ContinueAfterCard,
					tt.wantContinueAfter,
				)
			}
		})
	}
}
