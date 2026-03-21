package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// REST handles HTTP API calls to the go-risk-it server.
type REST struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewREST(baseURL, token string) *REST {
	return &REST{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}
}

func (r *REST) do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader

	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}

		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, r.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)

	return r.client.Do(req)
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
		return fmt.Errorf("advance: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("advance: status %d: %s", resp.StatusCode, body)
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
		return fmt.Errorf("%s: %w", moveType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s: status %d: %s", moveType, resp.StatusCode, body)
	}

	return nil
}
