package management

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/lobby/signals"
)

type UserLobbies struct {
	Owned    []sqlc.GetOwnedLobbiesRow
	Joined   []sqlc.GetJoinedLobbiesRow
	Joinable []sqlc.GetJoinableLobbiesRow
}

type Service interface {
	JoinLobby(ctx ctx.LobbyContext, name string) error
	GetUserLobbies(ctx ctx.UserContext) (*UserLobbies, error)
}

type ServiceImpl struct {
	querier                 db.Querier
	lobbyStateChangedSignal signals.LobbyStateChangedSignal
}

var _ Service = (*ServiceImpl)(nil)

func NewService(
	querier db.Querier,
	lobbyStateChangedSignal signals.LobbyStateChangedSignal,
) *ServiceImpl {
	return &ServiceImpl{
		querier:                 querier,
		lobbyStateChangedSignal: lobbyStateChangedSignal,
	}
}

func (s *ServiceImpl) JoinLobby(ctx ctx.LobbyContext, name string) error {
	if _, err := s.querier.ExecuteInTransaction(ctx, func(qtx db.Querier) (interface{}, error) {
		return nil, s.JoinLobbyQ(ctx, qtx, name)
	}); err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	go s.lobbyStateChangedSignal.Emit(ctx, signals.LobbyStateChangedData{})

	return nil
}

func (s *ServiceImpl) JoinLobbyQ(
	ctx ctx.LobbyContext,
	querier db.Querier,
	name string,
) error {
	ctx.Log().Infow("joining lobby")

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: ctx.LobbyID(),
		UserID:  ctx.UserID(),
		Name:    name,
	})
	if err != nil {
		return fmt.Errorf("failed to insert participant: %w", err)
	}

	ctx.Log().Infow("participant joined", "participant_id", participantID)

	return nil
}

func (s *ServiceImpl) GetUserLobbies(
	ctx ctx.UserContext,
) (*UserLobbies, error) {
	return s.GetUserLobbiesQ(ctx, s.querier)
}

func (s *ServiceImpl) GetUserLobbiesQ(
	ctx ctx.UserContext,
	querier db.Querier,
) (*UserLobbies, error) {
	ctx.Log().Infow("getting user lobbies")

	ownedLobbies, err := querier.GetOwnedLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get owned lobbies: %w", err)
	}

	ctx.Log().Infow("got owned lobbies", "lobbies", ownedLobbies)

	joinedLobbies, err := querier.GetJoinedLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joined lobbies: %w", err)
	}

	ctx.Log().Infow("got joined lobbies", "lobbies", joinedLobbies)

	joinableLobbies, err := querier.GetJoinableLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joinable lobbies: %w", err)
	}

	ctx.Log().Infow("got joinable lobbies", "lobbies", joinableLobbies)

	userLobbies := &UserLobbies{
		Owned:    ownedLobbies,
		Joined:   joinedLobbies,
		Joinable: joinableLobbies,
	}

	ctx.Log().Infow("got user lobbies", "lobbies", userLobbies)

	return userLobbies, nil
}
