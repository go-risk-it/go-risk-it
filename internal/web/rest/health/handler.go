package health

import (
	"net/http"
	"time"

	"github.com/go-risk-it/go-risk-it/internal/config"
	"github.com/go-risk-it/go-risk-it/internal/web/rest/route"
	"github.com/hellofresh/health-go/v5"
	"github.com/hellofresh/health-go/v5/checks/postgres"
)

type Handler interface {
	route.Route
}

type HandlerImpl struct {
	*health.Health
}

var _ Handler = (*HandlerImpl)(nil)

func New(databaseConfig config.DatabaseConfig) *HandlerImpl {
	health, err := health.New(
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
		panic(err)
	}

	return &HandlerImpl{
		Health: health,
	}
}

func (h *HandlerImpl) Pattern() string {
	return "/status"
}

func (h *HandlerImpl) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	h.Handler().ServeHTTP(writer, req)
}
