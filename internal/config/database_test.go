package config_test

import (
	"testing"

	"github.com/go-risk-it/go-risk-it/internal/config"
)

func TestDatabaseConfig_BuildConnectionString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		config     config.DatabaseConfig
		searchPath string
		want       string
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
			searchPath: "public",
			want:       "postgresql://user:pass@localhost:5432/testdb?search_path=public",
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
			searchPath: "public",
			want:       "postgresql://user:pass@localhost:5432/testdb?sslmode=disable&search_path=public",
		},
		{
			name: "connection with custom search path",
			config: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Name:     "testdb",
				User:     "user",
				Password: "pass",
			},
			searchPath: "custom_schema",
			want:       "postgresql://user:pass@localhost:5432/testdb?search_path=custom_schema",
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
			searchPath: "public",
			want:       "postgresql://user@domain:pass!123@db.example.com:5432/prod_db?search_path=public",
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
			searchPath: "public",
			want:       "postgresql://user:pass@localhost:54321/testdb?search_path=public",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := test.config.BuildConnectionString(test.searchPath)
			if got != test.want {
				t.Errorf("BuildConnectionString() = %v, want %v", got, test.want)
			}
		})
	}
}
