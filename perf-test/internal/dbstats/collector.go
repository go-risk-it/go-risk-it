package dbstats

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq" // Postgres driver.
)

// Collector reads database performance statistics from pg_stat_statements.
type Collector struct {
	db *sql.DB
}

// NewCollector creates a Collector connected to the given Postgres DSN.
func NewCollector(dsn string) (*Collector, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()

		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Collector{db: db}, nil
}

// Reset clears accumulated pg_stat_statements counters.
func (c *Collector) Reset(ctx context.Context) error {
	_, err := c.db.ExecContext(ctx, "SELECT pg_stat_statements_reset()")
	if err != nil {
		return fmt.Errorf("pg_stat_statements_reset: %w", err)
	}

	return nil
}

// Snapshot returns the top N queries by total execution time, plus connection count.
func (c *Collector) Snapshot(ctx context.Context, topN int) (StepDBStats, error) {
	var stats StepDBStats

	// Get top queries by total_exec_time.
	rows, err := c.db.QueryContext(ctx, `
		SELECT query, calls, total_exec_time, mean_exec_time, max_exec_time
		FROM pg_stat_statements
		WHERE userid = (SELECT usesysid FROM pg_user WHERE usename = current_user)
		ORDER BY total_exec_time DESC
		LIMIT $1
	`, topN)
	if err != nil {
		return stats, fmt.Errorf("query pg_stat_statements: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var qf QueryFingerprint
		if err := rows.Scan(&qf.Query, &qf.Calls, &qf.TotalTimeMs, &qf.MeanTimeMs, &qf.MaxTimeMs); err != nil {
			return stats, fmt.Errorf("scan row: %w", err)
		}

		stats.TopQueries = append(stats.TopQueries, qf)
		stats.TotalQueryTimeMs += qf.TotalTimeMs
	}

	if err := rows.Err(); err != nil {
		return stats, fmt.Errorf("rows iteration: %w", err)
	}

	// Get active connection count.
	err = c.db.QueryRowContext(ctx, `
		SELECT count(*) FROM pg_stat_activity
		WHERE state = 'active' AND pid != pg_backend_pid()
	`).Scan(&stats.ActiveConnections)
	if err != nil {
		return stats, fmt.Errorf("query pg_stat_activity: %w", err)
	}

	return stats, nil
}

// Close closes the database connection.
func (c *Collector) Close() error {
	return c.db.Close()
}
