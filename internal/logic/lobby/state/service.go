package state

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
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
	GetLobbyStateQ(ctx ctx.LobbyContext, querier db.Querier) (*Lobby, error)
}

type ServiceImpl struct {
	querier db.Querier
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
) *ServiceImpl {
	return &ServiceImpl{
		querier: querier,
	}
}

func (s *ServiceImpl) GetLobbyState(ctx ctx.LobbyContext) (*Lobby, error) {
	return s.GetLobbyStateQ(ctx, s.querier)
}

func (s *ServiceImpl) GetLobbyStateQ(ctx ctx.LobbyContext, querier db.Querier) (*Lobby, error) {
	lobby, err := querier.GetLobby(ctx, ctx.LobbyID())
	if err != nil {
		ctx.Log().Warnw("failed to get lobby", "err", err)

		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}

	if len(lobby) == 0 {
		return nil, errors.New("no participants in lobby")
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
