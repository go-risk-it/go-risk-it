package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Auth handles Supabase authentication.
type Auth struct {
	baseURL string
	anonKey string
	client  *http.Client
}

// AuthResult holds the result of a signup.
type AuthResult struct {
	UserID      string
	AccessToken string
}

func NewAuth(baseURL, anonKey string) *Auth {
	return &Auth{
		baseURL: baseURL,
		anonKey: anonKey,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type signupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signupResponse struct {
	AccessToken string `json:"accessToken"`
	User        struct {
		ID string `json:"id"`
	} `json:"user"`
}

// Signup creates a new user and returns the auth result.
func (a *Auth) Signup(email, password string) (*AuthResult, error) {
	ctx := context.Background()

	body, err := json.Marshal( //nolint:gosec // intentional for loadtest tool
		signupRequest{Email: email, Password: password},
	)
	if err != nil {
		return nil, fmt.Errorf("marshal signup: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		a.baseURL+"/auth/v1/signup",
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("create signup request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Apikey", a.anonKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("signup request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("signup failed: status %d", resp.StatusCode)
	}

	var result signupResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode signup response: %w", err)
	}

	return &AuthResult{
		UserID:      result.User.ID,
		AccessToken: result.AccessToken,
	}, nil
}
