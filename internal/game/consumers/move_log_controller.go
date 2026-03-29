package consumers

import (
	"errors"
	"fmt"
	"log/slog"

	game "github.com/go-risk-it/go-risk-it/internal/game/api"
	"github.com/go-risk-it/go-risk-it/internal/game/api/messaging"
	"github.com/go-risk-it/go-risk-it/internal/game/ctx"
	"github.com/go-risk-it/go-risk-it/internal/game/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/game/logic/move/orchestration/logging"
)

// MoveLogController translates between sqlc move log rows and messaging DTOs.
// It lives in consumers because its only callers are the publisher handlers
// in this package.
type MoveLogController struct {
	loggingService logging.Service
}

// NewMoveLogController creates a MoveLogController backed by the logging service.
func NewMoveLogController(loggingService logging.Service) *MoveLogController {
	return &MoveLogController{loggingService: loggingService}
}

// GetMoveLogs fetches recent move logs and converts them to messaging DTOs.
func (c *MoveLogController) GetMoveLogs(
	ctx ctx.GameContext,
	limit int64,
) (messaging.MoveHistory, error) {
	if limit < 0 {
		return messaging.MoveHistory{}, errors.New("limit must be positive")
	}

	if limit > 1000 {
		return messaging.MoveHistory{}, errors.New("limit must be less than 1000")
	}

	moveLogs, err := c.loggingService.GetMoveLogs(ctx, limit)
	if err != nil {
		return messaging.MoveHistory{}, fmt.Errorf("unable to get move history: %w", err)
	}

	convertedMoveLogs, err := convertMoveLogs(moveLogs)
	if err != nil {
		return messaging.MoveHistory{}, fmt.Errorf("unable to convert move logs: %w", err)
	}

	return messaging.MoveHistory{
		Moves: convertedMoveLogs,
	}, nil
}

// ConvertMoveLogs converts raw sqlc move log rows into messaging DTOs.
func (c *MoveLogController) ConvertMoveLogs(
	ctx ctx.GameContext,
	sqlcLogs []sqlc.GameMoveLog,
) (messaging.MoveHistory, error) {
	slog.DebugContext(ctx, "converting move logs")

	result := make([]messaging.MovePerformed, 0)

	for _, sqlcLog := range sqlcLogs {
		convertedSqlcLog, err := convertSqlcLog(ctx.UserID(), sqlcLog)
		if err != nil {
			return messaging.MoveHistory{}, fmt.Errorf("failed to convert move log: %w", err)
		}

		result = append(result, convertedSqlcLog)
	}

	slog.DebugContext(ctx, "converted move logs", "moves", result)

	return messaging.MoveHistory{
		Moves: result,
	}, nil
}

func convertMoveLogs(moveLogs []sqlc.GetMoveLogsRow) ([]messaging.MovePerformed, error) {
	result := make([]messaging.MovePerformed, 0)

	for _, m := range moveLogs {
		converted, err := convertMoveLog(m)
		if err != nil {
			return nil, fmt.Errorf("unable to convert move log: %w", err)
		}

		result = append(result, converted)
	}

	return result, nil
}

func convertSqlcLog(userID string, sqlcLog sqlc.GameMoveLog) (messaging.MovePerformed, error) {
	phase, err := convertPhase(sqlcLog.Phase)
	if err != nil {
		return messaging.MovePerformed{}, fmt.Errorf("unable to convert phase: %w", err)
	}

	return messaging.MovePerformed{
		Phase:   phase,
		Move:    sqlcLog.MoveData,
		Result:  sqlcLog.Result,
		Created: sqlcLog.Created.Time,
		UserID:  userID,
	}, nil
}

func convertMoveLog(moveLog sqlc.GetMoveLogsRow) (messaging.MovePerformed, error) {
	phase, err := convertPhase(moveLog.Phase)
	if err != nil {
		return messaging.MovePerformed{}, fmt.Errorf("unable to convert phase: %w", err)
	}

	return messaging.MovePerformed{
		Phase:   phase,
		Move:    moveLog.MoveData,
		Result:  moveLog.Result,
		Created: moveLog.Created.Time,
		UserID:  moveLog.UserID,
	}, nil
}

func convertPhase(phaseType sqlc.GamePhaseType) (game.PhaseType, error) {
	switch phaseType {
	case sqlc.GamePhaseTypeCARDS:
		return game.Cards, nil
	case sqlc.GamePhaseTypeDEPLOY:
		return game.Deploy, nil
	case sqlc.GamePhaseTypeATTACK:
		return game.Attack, nil
	case sqlc.GamePhaseTypeCONQUER:
		return game.Conquer, nil
	case sqlc.GamePhaseTypeREINFORCE:
		return game.Reinforce, nil
	default:
		return "", fmt.Errorf("invalid phase type: %s", phaseType)
	}
}
