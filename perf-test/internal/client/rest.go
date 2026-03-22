package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/perf-test/internal/metrics"
)

const (
	maxRetries    = 3
	baseBackoff   = 100 * time.Millisecond
	backoffFactor = 2
)

// ConflictError is returned on HTTP 409 (stale state).
type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return e.Message
}

// REST handles HTTP API calls to the go-risk-it server.
type REST struct {
	baseURL   string
	token     string
	client    *http.Client
	collector *metrics.Collector
}

// NewREST creates a REST client. If transport is non-nil it is used for
// connection pooling; otherwise a default transport is created.
func NewREST(
	baseURL, token string,
	transport *http.Transport,
	collector *metrics.Collector,
) *REST {
	var httpClient *http.Client
	if transport != nil {
		httpClient = &http.Client{Transport: transport}
	} else {
		httpClient = &http.Client{}
	}

	return &REST{
		baseURL:   baseURL,
		token:     token,
		client:    httpClient,
		collector: collector,
	}
}

func (r *REST) do(method, path string, body any) (*http.Response, error) {
	// Marshal body once so it can be replayed on retry.
	var bodyBytes []byte

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}

		bodyBytes = data
	}

	backoff := baseBackoff

	for attempt := range maxRetries {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequest(method, r.baseURL+path, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+r.token)

		resp, err := r.client.Do(req)
		if err != nil {
			if classified := classifyNetError(err); classified != nil && attempt < maxRetries-1 {
				log.Printf("retrying %s %s (attempt %d/%d): %v",
					method, path, attempt+2, maxRetries, err)

				if r.collector != nil {
					r.collector.RecordRetry()
				}

				time.Sleep(backoff)
				backoff *= backoffFactor

				continue
			}

			return nil, err
		}

		// Record every HTTP response status.
		if r.collector != nil {
			r.collector.RecordHTTPStatus(resp.StatusCode)
		}

		// Check for retryable HTTP status.
		if attempt < maxRetries-1 {
			if transient := classifyHTTPStatus(resp.StatusCode, nil); transient != nil {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				log.Printf("retrying %s %s (attempt %d/%d): HTTP %d: %s",
					method, path, attempt+2, maxRetries, resp.StatusCode, body)

				if r.collector != nil {
					r.collector.RecordRetry()
				}

				time.Sleep(backoff)
				backoff *= backoffFactor

				continue
			}
		}

		return resp, nil
	}

	// Unreachable, but the compiler needs it.
	return nil, errors.New("retry loop exhausted")
}

// Game creation types.
type CreateGamePlayer struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

type CreateGameRequest struct {
	Players []CreateGamePlayer `json:"players"`
}

type CreateGameResponse struct {
	GameID int64 `json:"gameId"`
}

func (r *REST) CreateGame(req CreateGameRequest) (int64, error) {
	resp, err := r.do(http.MethodPost, "/api/v1/games", req)
	if err != nil {
		return 0, fmt.Errorf("create game: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return 0, fmt.Errorf("create game: status %d: %s", resp.StatusCode, body)
	}

	var result CreateGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode create game response: %w", err)
	}

	return result.GameID, nil
}

// Move types matching the server's request types exactly.

type DeployMove struct {
	RegionID      string `json:"regionId"`
	CurrentTroops int64  `json:"currentTroops"`
	DesiredTroops int64  `json:"desiredTroops"`
}

func (r *REST) Deploy(gameID int64, move DeployMove) error {
	return r.doMove(gameID, "deployments", move)
}

type AttackMove struct {
	SourceRegionID  string `json:"sourceRegionId"`
	TargetRegionID  string `json:"targetRegionId"`
	TroopsInSource  int64  `json:"troopsInSource"`
	TroopsInTarget  int64  `json:"troopsInTarget"`
	AttackingTroops int64  `json:"attackingTroops"`
}

func (r *REST) Attack(gameID int64, move AttackMove) error {
	return r.doMove(gameID, "attacks", move)
}

type ConquerMove struct {
	Troops int64 `json:"troops"`
}

func (r *REST) Conquer(gameID int64, move ConquerMove) error {
	return r.doMove(gameID, "conquers", move)
}

type ReinforceMove struct {
	SourceRegionID string `json:"sourceRegionId"`
	TargetRegionID string `json:"targetRegionId"`
	TroopsInSource int64  `json:"troopsInSource"`
	TroopsInTarget int64  `json:"troopsInTarget"`
	MovingTroops   int64  `json:"movingTroops"`
}

func (r *REST) Reinforce(gameID int64, move ReinforceMove) error {
	return r.doMove(gameID, "reinforcements", move)
}

type CardCombination struct {
	CardIDs []int64 `json:"cardIds"`
}

type CardsMove struct {
	Combinations []CardCombination `json:"combinations"`
}

func (r *REST) PlayCards(gameID int64, move CardsMove) error {
	return r.doMove(gameID, "cards", move)
}

type Advancement struct {
	CurrentPhase string `json:"currentPhase"`
}

func (r *REST) Advance(gameID int64, currentPhase string) error {
	resp, err := r.do(
		http.MethodPost,
		fmt.Sprintf("/api/v1/games/%d/advancements", gameID),
		Advancement{CurrentPhase: currentPhase},
	)
	if err != nil {
		if classified := classifyNetError(err); classified != nil {
			return fmt.Errorf("advance: %w", classified)
		}

		return fmt.Errorf("advance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("advance: status %d: %s", resp.StatusCode, body)

		if resp.StatusCode == http.StatusConflict {
			return &ConflictError{Message: msg}
		}

		if transient := classifyHTTPStatus(resp.StatusCode, fmt.Errorf("%s", msg)); transient != nil {
			return transient
		}

		return fmt.Errorf("%s", msg)
	}

	return nil
}

func (r *REST) doMove(gameID int64, moveType string, move any) error {
	resp, err := r.do(
		http.MethodPost,
		fmt.Sprintf("/api/v1/games/%d/moves/%s", gameID, moveType),
		move,
	)
	if err != nil {
		if classified := classifyNetError(err); classified != nil {
			return fmt.Errorf("%s: %w", moveType, classified)
		}

		return fmt.Errorf("%s: %w", moveType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		msg := fmt.Sprintf("%s: status %d: %s", moveType, resp.StatusCode, body)

		if resp.StatusCode == http.StatusConflict {
			return &ConflictError{Message: msg}
		}

		if resp.StatusCode == http.StatusBadRequest {
			return &StaleStateError{Message: msg}
		}

		if transient := classifyHTTPStatus(resp.StatusCode, fmt.Errorf("%s", msg)); transient != nil {
			return transient
		}

		return fmt.Errorf("%s", msg)
	}

	return nil
}
