package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/kernel/config"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/hellofresh/health-go/v5"
	"github.com/hellofresh/health-go/v5/checks/postgres"
)

func New(databaseConfig config.DatabaseConfig) (*route.Route, error) {
	healthCheck, err := health.New(
		health.WithComponent(health.Component{
			Name:    "go-risk-it",
			Version: "1.0.0",
		}),
		health.WithSystemInfo(),
		health.WithChecks(
			health.Config{
				Name:      "postgres",
				Timeout:   5 * time.Second,
				SkipOnErr: false,
				Check: postgres.New(postgres.Config{
					DSN: databaseConfig.BuildConnectionString(),
				}),
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create health handler: %w", err)
	}

	return route.Public("GET /status", func(w http.ResponseWriter, r *http.Request) error {
		healthCheck.Handler().ServeHTTP(w, r)

		return nil
	}), nil
}
