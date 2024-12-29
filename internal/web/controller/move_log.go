package controller

import (
	"errors"
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/api/game"
	"github.com/go-risk-it/go-risk-it/internal/api/game/messaging"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/sqlc"
	"github.com/go-risk-it/go-risk-it/internal/logic/game/move/orchestration/logging"
)

type MoveLogController interface {
	ConvertMoveLogs(ctx ctx.GameContext, sqlcLogs []sqlc.MoveLog) (messaging.MoveHistory, error)
	GetMoveLogs(ctx ctx.GameContext, limit int64) (messaging.MoveHistory, error)
}

type MoveLogControllerImpl struct {
	loggingService logging.Service
}

var _ MoveLogController = (*MoveLogControllerImpl)(nil)

func NewMoveLogController(loggingService logging.Service) *MoveLogControllerImpl {
	return &MoveLogControllerImpl{loggingService: loggingService}
}

func (c *MoveLogControllerImpl) GetMoveLogs(
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

func (c *MoveLogControllerImpl) ConvertMoveLogs(
	ctx ctx.GameContext,
	sqlcLogs []sqlc.MoveLog,
) (messaging.MoveHistory, error) {
	ctx.Log().Debug("converting move logs")

	result := make([]messaging.MovePerformed, 0)

	for _, sqlcLog := range sqlcLogs {
		convertedSqlcLog, err := convertSqlcLog(ctx.UserID(), sqlcLog)
		if err != nil {
			return messaging.MoveHistory{}, fmt.Errorf("failed to convert move log: %w", err)
		}

		result = append(result, convertedSqlcLog)
	}

	ctx.Log().Debugf("converted move logs: %s", result)

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

func convertSqlcLog(userID string, sqlcLog sqlc.MoveLog) (messaging.MovePerformed, error) {
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

func convertPhase(phaseType sqlc.PhaseType) (game.PhaseType, error) {
	switch phaseType {
	case sqlc.PhaseTypeCARDS:
		return game.Cards, nil
	case sqlc.PhaseTypeDEPLOY:
		return game.Deploy, nil
	case sqlc.PhaseTypeATTACK:
		return game.Attack, nil
	case sqlc.PhaseTypeCONQUER:
		return game.Conquer, nil
	case sqlc.PhaseTypeREINFORCE:
		return game.Reinforce, nil
	default:
		return "", fmt.Errorf("invalid phase type: %s", phaseType)
	}
}
