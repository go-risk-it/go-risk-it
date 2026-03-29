//go:build invariant

package invariant

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	gamedb "github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	game "github.com/go-risk-it/go-risk-it/internal/game/logic"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/advancement"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/board"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/creation"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/conquer"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/deploy"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/player"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/state"
	"github.com/go-risk-it/go-risk-it/internal/game/rand"
	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	"github.com/go-risk-it/go-risk-it/internal/kernel/metrics"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/fx"
)

// GameHandle holds references for a created game.
type GameHandle struct {
	GameID  int64
	Players []player.Player
}

// Harness owns the test infrastructure:
// testcontainer, pool, fx app, and all service references.
type Harness struct {
	CreationService       creation.Service
	DeployOrchestrator    orchestration.DeployOrchestrator
	AttackOrchestrator    orchestration.AttackOrchestrator
	ConquerOrchestrator   orchestration.ConquerOrchestrator
	ReinforceOrchestrator orchestration.ReinforceOrchestrator
	CardsOrchestrator     orchestration.CardsOrchestrator
	AttackAdvancer        advancement.AttackAdvancer
	CardsAdvancer         advancement.CardsAdvancer
	ReinforceAdvancer     advancement.ReinforceAdvancer
	StateService          state.Service
	DeployService         deploy.Service
	ConquerService        conquer.Service
	BoardService          board.Service
	Querier               gamedb.Querier

	pool *pgxpool.Pool
}

// NewHarness creates a Harness backed by a testcontainer Postgres.
// It runs migrations, builds the fx app, and populates all
// service references.
func NewHarness() *Harness {
	dbPool, err := setupDatabase()
	if err != nil {
		panic(fmt.Sprintf("failed to setup database: %v", err))
	}

	harness := &Harness{
		pool: dbPool,
	}

	buildFxApp(harness, dbPool)

	return harness
}

func buildFxApp(
	harness *Harness,
	dbPool *pgxpool.Pool,
) {
	reader := metric.NewManualReader()
	provider := metric.NewMeterProvider(metric.WithReader(reader))
	meter := provider.Meter("invariant-test")

	testMetrics, err := metrics.NewInfraMetrics(meter)
	if err != nil {
		panic(fmt.Sprintf("failed to create test metrics: %v", err))
	}

	testKoanf := buildTestKoanf()

	app := fx.New(
		fx.NopLogger,
		fx.Provide(func() gamedb.DB { return dbPool }),
		fx.Provide(func() sqlc.DBTX { return dbPool }),
		fx.Provide(gamedb.New),
		game.Module,
		fx.Supply(testKoanf),
		rand.Module,
		fx.Supply(testMetrics),
		fx.Populate(
			&harness.CreationService,
			&harness.DeployOrchestrator,
			&harness.AttackOrchestrator,
			&harness.ConquerOrchestrator,
			&harness.ReinforceOrchestrator,
			&harness.CardsOrchestrator,
			&harness.AttackAdvancer,
			&harness.CardsAdvancer,
			&harness.ReinforceAdvancer,
			&harness.StateService,
			&harness.DeployService,
			&harness.ConquerService,
			&harness.BoardService,
			&harness.Querier,
		),
	)

	if err := app.Err(); err != nil {
		panic(fmt.Sprintf("fx app failed: %v", err))
	}
}

// CreateGame creates a new game with the given number of players.
func (h *Harness) CreateGame(
	tb testing.TB,
	nPlayers int,
) *GameHandle {
	tb.Helper()

	regions, err := h.BoardService.GetBoardRegions(h.logCtx())
	if err != nil {
		tb.Fatalf("failed to get board regions: %v", err)
	}

	players := make([]player.Player, nPlayers)
	for i := range nPlayers {
		players[i] = player.Player{
			UserID: fmt.Sprintf("user-%d", i),
			Name:   fmt.Sprintf("Player %d", i),
		}
	}

	userCtx := h.userCtx(players[0].UserID)

	gameID, err := h.CreationService.CreateGame(
		userCtx, regions, players,
	)
	if err != nil {
		tb.Fatalf("failed to create game: %v", err)
	}

	return &GameHandle{GameID: gameID, Players: players}
}

// GameCtx builds a full GameContext for the given game and user.
func (h *Harness) GameCtx(
	gameID int64,
	userID string,
) ctx.GameContext {
	span := noop.Span{}
	traceCtx := kernelctx.WithSpan(context.Background(), span)
	userCtx := kernelctx.WithUserID(traceCtx, userID)

	return ctx.WithGameID(userCtx, gameID)
}

// Close shuts down the connection pool.
func (h *Harness) Close() {
	h.pool.Close()
}

func (h *Harness) logCtx() context.Context {
	return context.Background()
}

func (h *Harness) userCtx(userID string) kernelctx.UserContext {
	span := noop.Span{}
	traceCtx := kernelctx.WithSpan(context.Background(), span)

	return kernelctx.WithUserID(traceCtx, userID)
}

func setupDatabase() (*pgxpool.Pool, error) {
	bgCtx := context.Background()

	connStr, err := startPostgres(bgCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres: %w", err)
	}

	if err := runMigrations(connStr); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	pool, err := pgxpool.New(bgCtx, connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	return pool, nil
}

func startPostgres(bgCtx context.Context) (string, error) {
	container, err := postgres.Run(bgCtx,
		"docker.io/postgres:latest",
		postgres.WithDatabase("risk-it"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog(
				"database system is ready to accept connections",
			).
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return "", fmt.Errorf(
			"failed to run testcontainer: %w", err,
		)
	}

	connStr, err := container.ConnectionString(
		bgCtx, "sslmode=disable",
	)
	if err != nil {
		return "", fmt.Errorf(
			"failed to get connection string: %w", err,
		)
	}

	return connStr, nil
}

func runMigrations(connStr string) error {
	_, callerPath, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("failed to get caller path")
	}

	migrationDir := filepath.Join(
		filepath.Dir(callerPath),
		"..", "..", "game", "data", "sqlc", "migrations",
	)

	mig, err := migrate.New("file://"+migrationDir, connStr)
	if err != nil {
		return fmt.Errorf(
			"failed to create migrate instance: %w", err,
		)
	}

	defer mig.Close()

	if err := mig.Up(); err != nil &&
		!errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func buildTestKoanf() *koanf.Koanf {
	raw := []byte(`
game:
  dice:
    roll_strategy: attacker_always_wins
  regionassignment:
    assignment_strategy: sequential
  history:
    size: 50
`)

	k := koanf.New(".")

	if err := k.Load(rawbytes.Provider(raw), yaml.Parser()); err != nil {
		panic(fmt.Sprintf("failed to load test koanf: %v", err))
	}

	return k
}
