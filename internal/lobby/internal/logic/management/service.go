package management

import (
	"errors"
	"fmt"

	eventbus "github.com/go-risk-it/go-risk-it/internal/kernel/bus"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	dbutil "github.com/go-risk-it/go-risk-it/internal/kernel/data"
	domainerrors "github.com/go-risk-it/go-risk-it/internal/kernel/errors"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/go-risk-it/go-risk-it/internal/kernel/observe"
	"github.com/go-risk-it/go-risk-it/internal/lobby/api/snapshot"
	"github.com/go-risk-it/go-risk-it/internal/lobby/ctx"
	lobbyevt "github.com/go-risk-it/go-risk-it/internal/lobby/events"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/data/db"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/lobby/internal/logic/state"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserLobbies struct {
	Owned    []sqlc.GetOwnedLobbiesRow
	Joined   []sqlc.GetJoinedLobbiesRow
	Joinable []sqlc.GetJoinableLobbiesRow
}

type Service interface {
	JoinLobby(ctx ctx.LobbyContext, name string) error
	JoinLobbyWithQuerier(ctx ctx.LobbyContext, querier db.Querier, name string) error
	GetUserLobbies(ctx kernelctx.UserContext) (*UserLobbies, error)
	GetUserLobbiesWithQuerier(ctx kernelctx.UserContext, querier db.Querier) (*UserLobbies, error)
}

type service struct {
	querier      db.Querier
	bus          eventbus.Publisher
	metrics      *metrics.StateMetrics
	stateService state.Service
}

var _ Service = (*service)(nil)

func NewService(
	querier db.Querier,
	bus eventbus.Publisher,
	m *metrics.StateMetrics,
	stateService state.Service,
) Service {
	return &service{
		querier:      querier,
		bus:          bus,
		metrics:      m,
		stateService: stateService,
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

	snap, err := s.buildLobbySnapshot(ctx)
	if err != nil {
		observe.Warn(ctx, "failed to build lobby snapshot for event, emitting without snapshot")

		snap = nil
	}

	s.bus.Emit(ctx, lobbyevt.NewLobbyStateChanged(ctx.LobbyID(), ctx.UserID(), snap))

	return nil
}

func (s *service) JoinLobbyWithQuerier(
	ctx ctx.LobbyContext,
	querier db.Querier,
	name string,
) error {
	_, err := querier.InsertParticipant(ctx, sqlc.InsertParticipantParams{
		LobbyID: ctx.LobbyID(),
		UserID:  ctx.UserID(),
		Name:    name,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.NewConflictError("already joined this lobby")
		}

		return fmt.Errorf("failed to insert participant: %w", err)
	}

	return nil
}

func (s *service) GetUserLobbies(
	ctx kernelctx.UserContext,
) (*UserLobbies, error) {
	return s.GetUserLobbiesWithQuerier(ctx, s.querier)
}

func (s *service) GetUserLobbiesWithQuerier(
	ctx kernelctx.UserContext,
	querier db.Querier,
) (*UserLobbies, error) {
	return observe.Span(ctx, "lobby.get_user_lobbies",
		func(ctx kernelctx.UserContext) (*UserLobbies, error) {
			ownedLobbies, err := querier.GetOwnedLobbies(ctx, ctx.UserID())
			if err != nil {
				return nil, fmt.Errorf("failed to get owned lobbies: %w", err)
			}

			joinedLobbies, err := querier.GetJoinedLobbies(ctx, ctx.UserID())
			if err != nil {
				return nil, fmt.Errorf("failed to get joined lobbies: %w", err)
			}

			joinableLobbies, err := querier.GetJoinableLobbies(ctx, ctx.UserID())
			if err != nil {
				return nil, fmt.Errorf("failed to get joinable lobbies: %w", err)
			}

			return &UserLobbies{
				Owned:    ownedLobbies,
				Joined:   joinedLobbies,
				Joinable: joinableLobbies,
			}, nil
		},
	)
}

func (s *service) buildLobbySnapshot(ctx ctx.LobbyContext) (*snapshot.LobbySnapshot, error) {
	lobby, err := s.stateService.GetLobbyState(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby state: %w", err)
	}

	participants := make([]snapshot.Participant, 0, len(lobby.Participants))
	for _, p := range lobby.Participants {
		participants = append(participants, snapshot.Participant{
			UserID: p.UserID,
		})
	}

	return &snapshot.LobbySnapshot{
		ID:           lobby.ID,
		Participants: participants,
	}, nil
}
