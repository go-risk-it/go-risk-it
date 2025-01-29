package testonly

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/game/pool"
)

type Service interface {
	TruncateTables(ctx ctx.LogContext) error
}

type ServiceImpl struct {
	pool     pool.DB
	dbConfig config.DatabaseConfig
	tables   []string
}

var _ Service = (*ServiceImpl)(nil)

func NewService(pool pool.DB, dbConfig config.DatabaseConfig) *ServiceImpl {
	tables := []string{
		"player",
		"game",
		"region",
		"card",
		"mission",
		"deploy_phase",
		"conquer_phase",
		"phase",
		"move_log",
	}

	return &ServiceImpl{pool: pool, dbConfig: dbConfig, tables: tables}
}

func (s *ServiceImpl) TruncateTables(ctx ctx.LogContext) error {
	ctx.Log().Infow("Truncating tables", "tables", s.tables)

	for _, table := range s.tables {
		ctx.Log().Infow("Truncating table", "table", table)

		_, err := s.pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	ctx.Log().Infow("Truncated tables", "tables", s.tables)

	return nil
}
