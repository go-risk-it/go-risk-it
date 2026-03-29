package management

import (
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/data/lobby/db"
	"github.com/go-risk-it/go-risk-it/internal/data/lobby/sqlc"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/events/lobby"
	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	"github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
)

type UserLobbies struct {
	Owned    []sqlc.GetOwnedLobbiesRow
	Joined   []sqlc.GetJoinedLobbiesRow
	Joinable []sqlc.GetJoinableLobbiesRow
}

type Service interface {
	JoinLobby(ctx ctx.LobbyContext, name string) error
	JoinLobbyWithQuerier(ctx ctx.LobbyContext, querier db.Querier, name string) error
	GetUserLobbies(ctx ctx.UserContext) (*UserLobbies, error)
	GetUserLobbiesWithQuerier(ctx ctx.UserContext, querier db.Querier) (*UserLobbies, error)
}

type service struct {
	querier db.Querier
	bus     eventbus.Bus
	metrics *metrics.Metrics
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	bus eventbus.Bus,
	m *metrics.Metrics,
) Service {
	return &service{
		querier: querier,
		bus:     bus,
		metrics: m,
	}
}

func (s *service) JoinLobby(ctx ctx.LobbyContext, name string) error {
	if _, err := dbutil.InTransaction(
		s.querier,
		ctx,
		s.metrics,
		func(qtx db.Querier) (struct{}, error) {
			return struct{}{}, s.JoinLobbyWithQuerier(ctx, qtx, name)
		}); err != nil {
		return fmt.Errorf("failed to join lobby: %w", err)
	}

	s.bus.Emit(ctx, lobbyevt.NewLobbyStateChanged(ctx.LobbyID(), ctx.UserID()))

	return nil
}

func (s *service) JoinLobbyWithQuerier(
	ctx ctx.LobbyContext,
	querier db.Querier,
	name string,
) error {
	slog.InfoContext(ctx, "joining lobby")

	participantID, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: ctx.LobbyID(),
		UserID:  ctx.UserID(),
		Name:    name,
	})
	if err != nil {
		return fmt.Errorf("failed to insert participant: %w", err)
	}

	slog.InfoContext(ctx, "participant joined", "participant_id", participantID)

	return nil
}

func (s *service) GetUserLobbies(
	ctx ctx.UserContext,
) (*UserLobbies, error) {
	return s.GetUserLobbiesWithQuerier(ctx, s.querier)
}

func (s *service) GetUserLobbiesWithQuerier(
	ctx ctx.UserContext,
	querier db.Querier,
) (*UserLobbies, error) {
	slog.InfoContext(ctx, "getting user lobbies")

	ownedLobbies, err := querier.GetOwnedLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get owned lobbies: %w", err)
	}

	slog.InfoContext(ctx, "got owned lobbies", "lobbies", ownedLobbies)

	joinedLobbies, err := querier.GetJoinedLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joined lobbies: %w", err)
	}

	slog.InfoContext(ctx, "got joined lobbies", "lobbies", joinedLobbies)

	joinableLobbies, err := querier.GetJoinableLobbies(ctx, ctx.UserID())
	if err != nil {
		return nil, fmt.Errorf("failed to get joinable lobbies: %w", err)
	}

	slog.InfoContext(ctx, "got joinable lobbies", "lobbies", joinableLobbies)

	userLobbies := &UserLobbies{
		Owned:    ownedLobbies,
		Joined:   joinedLobbies,
		Joinable: joinableLobbies,
	}

	slog.InfoContext(ctx, "got user lobbies", "lobbies", userLobbies)

	return userLobbies, nil
}
