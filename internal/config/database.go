package config

import (
	"fmt"
	"net"
	"strconv"
)

type DatabaseConfig struct {
	Host       string `koanf:"host"`
	Port       int    `koanf:"port"`
	Name       string `koanf:"name"`
	User       string `koanf:"user"`
	Password   string `koanf:"password"`
	DisableSSL bool   `koanf:"disable_ssl"`
}

func (c *DatabaseConfig) BuildConnectionString() string {
	hostPort := net.JoinHostPort(c.Host, strconv.Itoa(c.Port))

	result := fmt.Sprintf(
		"postgresql://%s:%s@%s/%s",
		c.User,
		c.Password,
		hostPort,
		c.Name,
	)

	if c.DisableSSL {
		result += "?sslmode=disable"
	}

	return result
}
