package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type DatabaseConfig struct {
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	Name       string `koanf:"name"`
	User       string `koanf:"user"`
	Password   string `koanf:"password"`
	DisableSSL bool   `koanf:"disable_ssl"`
}

func (c *DatabaseConfig) BuildConnectionString(searchPath string) string {
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	result := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		c.User,
		c.Password,
		hostPort,
		c.Name,
	)

	params := make([]string, 0, 2)

	if c.DisableSSL {
		params = append(params, "sslmode=disable")
	}

	params = append(params, "search_path="+searchPath)

	if len(params) > 0 {
		result += "?" + strings.Join(params, "&")
	}

	return result
}
