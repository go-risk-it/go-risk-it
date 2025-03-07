package config_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/config"
)

func TestDatabaseConfig_BuildConnectionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config config.DatabaseConfig
		want   string
	}{
		{
			name: "basic connection string",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				User:     "user",
				Password: "pass",
			},
			want: "postgresql://user:pass@localhost:5432/testdb",
		},
		{
			name: "connection with SSL disabled",
			config: config.DatabaseConfig{
				Host:       "localhost",
				Port:       5432,
				Name:       "testdb",
				User:       "user",
				Password:   "pass",
				DisableSSL: true,
			},
			want: "postgresql://user:pass@localhost:5432/testdb?sslmode=disable&search_path=public",
		},
		{
			name: "connection with special characters",
			config: config.DatabaseConfig{
				Host:     "db.example.com",
				Port:     5432,
				Name:     "prod_db",
				User:     "user@domain",
				Password: "pass!123",
			},
			want: "postgresql://user@domain:pass!123@db.example.com:5432/prod_db",
		},
		{
			name: "connection with non-standard port",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     54321,
				Name:     "testdb",
				User:     "user",
				Password: "pass",
			},
			want: "postgresql://user:pass@localhost:54321/testdb",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.config.BuildConnectionString()
			if got != test.want {
				t.Errorf("BuildConnectionString() = %v, want %v", got, test.want)
			}
		})
	}
}
