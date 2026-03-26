package state

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/logic/errors"
)

type Participant struct {
	UserID string
}

type Lobby struct {
	ID           int64
	Participants []Participant
}

type Service interface {
	GetLobbyState(ctx ctx.LobbyContext) (*Lobby, error)
	GetLobbyStateWithQuerier(ctx ctx.LobbyContext, querier db.Querier) (*Lobby, error)
}

type service struct {
	querier db.Querier
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
) Service {
	return &service{
		querier: querier,
	}
}

func (s *service) GetLobbyState(ctx ctx.LobbyContext) (*Lobby, error) {
	return s.GetLobbyStateWithQuerier(ctx, s.querier)
}

func (s *service) GetLobbyStateWithQuerier(
	ctx ctx.LobbyContext,
	querier db.Querier,
) (*Lobby, error) {
	lobby, err := querier.GetLobby(ctx, ctx.LobbyID())
	if err != nil {
		slog.WarnContext(ctx, "failed to get lobby", "error", err)

		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	if len(lobby) == 0 {
		return nil, domainerrors.NewNotFoundError("no participants in lobby")
	}

	participants := make([]Participant, 0)
	for _, p := range lobby {
		participants = append(participants, Participant{
			UserID: p.UserID,
		})
	}

	return &Lobby{
		ID:           lobby[0].ID,
		Participants: participants,
	}, nil
}
