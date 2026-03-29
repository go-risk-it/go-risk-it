package testonly

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-risk-it/go-risk-it/internal/game/data/db"
	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/jackc/pgx/v5"
)

type Service interface {
	TruncateTables(ctx context.Context) error
	SetupNearWin(ctx context.Context, gameID int64) error
}

type service struct {
	pool     db.DB
	dbConfig config.DatabaseConfig
	tables   []string
}

var _ Service = (*service)(nil)

func NewService(pool db.DB, dbConfig config.DatabaseConfig) Service {
	tables := []string{
		"game.card",
		"game.conquer_phase",
		"game.deploy_phase",
		"game.eliminate_player_mission",
		"game.game",
		"game.mission",
		"game.move_log",
		"game.phase",
		"game.player",
		"game.phase",
		"game.region",
		"game.two_continents_mission",
		"game.two_continents_plus_one_mission",
		"lobby.lobby",
		"lobby.participant",
	}

	return &service{pool: pool, dbConfig: dbConfig, tables: tables}
}

func (s *service) TruncateTables(ctx context.Context) error {
	slog.InfoContext(ctx, "Truncating tables", "tables", s.tables)

	for _, table := range s.tables {
		slog.InfoContext(ctx, "Truncating table", "table", table)

		_, err := s.pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	slog.InfoContext(ctx, "Truncated tables", "tables", s.tables)

	return nil
}

func (s *service) SetupNearWin(ctx context.Context, gameID int64) error {
	slog.InfoContext(ctx, "Setting up near-win state", "gameID", gameID)

	transaction, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		_ = transaction.Rollback(ctx)
	}()

	playerIDs, err := getPlayerIDs(ctx, transaction, gameID)
	if err != nil {
		return err
	}

	if len(playerIDs) < 3 {
		return fmt.Errorf("expected at least 3 players, got %d", len(playerIDs))
	}

	winner := playerIDs[0] // turn_index=0, will be the winner
	target := playerIDs[1] // target to eliminate

	if err := setupMissions(ctx, transaction, gameID, winner, target); err != nil {
		return err
	}

	if err := setupRegions(ctx, transaction, gameID, winner, target); err != nil {
		return err
	}

	if err := resetCardsAndPhases(ctx, transaction, gameID); err != nil {
		return err
	}

	if err := createDeployPhase(ctx, transaction, gameID); err != nil {
		return err
	}

	if err := transaction.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.InfoContext(ctx, "Near-win state set up successfully", "gameID", gameID)

	return nil
}

func getPlayerIDs(ctx context.Context, transaction pgx.Tx, gameID int64) ([]int64, error) {
	rows, err := transaction.Query(ctx,
		`SELECT id FROM game.player WHERE game_id = $1 ORDER BY turn_index`, gameID)
	if err != nil {
		return nil, fmt.Errorf("failed to get players: %w", err)
	}

	defer rows.Close()

	var playerIDs []int64

	for rows.Next() {
		var playerID int64
		if err := rows.Scan(&playerID); err != nil {
			return nil, fmt.Errorf("failed to scan player id: %w", err)
		}

		playerIDs = append(playerIDs, playerID)
	}

	return playerIDs, nil
}

func setupMissions(
	ctx context.Context,
	transaction pgx.Tx,
	gameID int64,
	winner int64,
	target int64,
) error {
	missionSubTables := []string{
		"game.eliminate_player_mission",
		"game.two_continents_mission",
		"game.two_continents_plus_one_mission",
	}

	for _, table := range missionSubTables {
		_, err := transaction.Exec(ctx, fmt.Sprintf(
			`DELETE FROM %s WHERE mission_id IN
				(SELECT id FROM game.mission WHERE player_id IN
					(SELECT id FROM game.player WHERE game_id = $1))`, table), gameID)
		if err != nil {
			return fmt.Errorf("failed to delete %s: %w", table, err)
		}
	}

	_, err := transaction.Exec(ctx,
		`DELETE FROM game.mission WHERE player_id IN
			(SELECT id FROM game.player WHERE game_id = $1)`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete missions: %w", err)
	}

	var missionID int64

	err = transaction.QueryRow(ctx,
		`INSERT INTO game.mission (player_id, type) VALUES ($1, 'ELIMINATE_PLAYER') RETURNING id`,
		winner).Scan(&missionID)
	if err != nil {
		return fmt.Errorf("failed to insert mission: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`INSERT INTO game.eliminate_player_mission (mission_id, target_player_id) VALUES ($1, $2)`,
		missionID, target)
	if err != nil {
		return fmt.Errorf("failed to insert eliminate_player_mission: %w", err)
	}

	return nil
}

func setupRegions(
	ctx context.Context,
	transaction pgx.Tx,
	gameID int64,
	winner int64,
	target int64,
) error {
	_, err := transaction.Exec(ctx,
		`UPDATE game.region SET player_id = $1, troops = 10
			WHERE id IN (SELECT r.id FROM game.region r
				JOIN game.player p ON r.player_id = p.id
				WHERE p.game_id = $2)`, winner, gameID)
	if err != nil {
		return fmt.Errorf("failed to assign all regions to winner: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`UPDATE game.region SET player_id = $1, troops = 1
			WHERE external_reference = 'eastern_australia'
			AND id IN (SELECT r.id FROM game.region r
				JOIN game.player p ON r.player_id = p.id
				WHERE p.game_id = $2)`, target, gameID)
	if err != nil {
		return fmt.Errorf("failed to assign eastern_australia to target: %w", err)
	}

	return nil
}

func resetCardsAndPhases(ctx context.Context, transaction pgx.Tx, gameID int64) error {
	_, err := transaction.Exec(ctx,
		`UPDATE game.card SET owner_id = NULL WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("failed to reset cards: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`UPDATE game.game SET current_phase_id = NULL WHERE id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("failed to clear current_phase_id: %w", err)
	}

	phaseSubTables := []string{"game.conquer_phase", "game.deploy_phase"}

	for _, table := range phaseSubTables {
		_, err = transaction.Exec(ctx, fmt.Sprintf(
			`DELETE FROM %s WHERE phase_id IN
				(SELECT id FROM game.phase WHERE game_id = $1)`, table), gameID)
		if err != nil {
			return fmt.Errorf("failed to delete %s: %w", table, err)
		}
	}

	_, err = transaction.Exec(ctx,
		`DELETE FROM game.phase WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete phases: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`DELETE FROM game.move_log WHERE game_id = $1`, gameID)
	if err != nil {
		return fmt.Errorf("failed to delete move logs: %w", err)
	}

	return nil
}

func createDeployPhase(ctx context.Context, transaction pgx.Tx, gameID int64) error {
	var phaseID int64

	err := transaction.QueryRow(ctx,
		`INSERT INTO game.phase (game_id, type, turn) VALUES ($1, 'DEPLOY', 0) RETURNING id`,
		gameID).Scan(&phaseID)
	if err != nil {
		return fmt.Errorf("failed to insert phase: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`INSERT INTO game.deploy_phase (phase_id, deployable_troops) VALUES ($1, 3)`, phaseID)
	if err != nil {
		return fmt.Errorf("failed to insert deploy_phase: %w", err)
	}

	_, err = transaction.Exec(ctx,
		`UPDATE game.game SET current_phase_id = $1, winner_player_id = NULL WHERE id = $2`,
		phaseID, gameID)
	if err != nil {
		return fmt.Errorf("failed to update game phase: %w", err)
	}

	return nil
}
