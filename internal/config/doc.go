// Package config loads application configuration from YAML files and
// environment variables using koanf. It reads a YAML file named after
// the ENVIRONMENT variable (defaulting to "component-test.yml"), then
// overlays any matching environment variables.
//
// Each configuration section (JWT, database, dice, region assignment,
// history, OpenTelemetry, server) is provided as a separate fx dependency,
// allowing consumers to declare only the config they need.
//
// # Layer
//
// Infrastructure — configuration loading with no internal dependencies.
package config
