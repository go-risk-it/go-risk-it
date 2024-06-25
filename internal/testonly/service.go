package testonly

import (
	"fmt"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/ctx"
	"github.com/go-risk-it/go-risk-it/internal/data/pool"
	"go.uber.org/zap"
)

type Service interface {
	TruncateTables(ctx ctx.LogContext) error
}

type ServiceImpl struct {
	log      *zap.SugaredLogger
	pool     pool.DB
	dbConfig config.DatabaseConfig
	tables   []string
}

func NewService(
	log *zap.SugaredLogger,
	pool pool.DB,
	dbConfig config.DatabaseConfig,
) *ServiceImpl {
	tables := []string{
		"player",
		"game",
		"region",
		"card",
		"mission",
	}

	return &ServiceImpl{log: log, pool: pool, dbConfig: dbConfig, tables: tables}
}

func (s *ServiceImpl) TruncateTables(ctx ctx.LogContext) error {
	s.log.Infow("Truncating tables", "tables", s.tables)

	for _, table := range s.tables {
		s.log.Infow("Truncating table", "table", table)

		_, err := s.pool.Exec(ctx, fmt.Sprintf("TRUNCATE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	s.log.Infow("Truncated tables", "tables", s.tables)

	return nil
}
