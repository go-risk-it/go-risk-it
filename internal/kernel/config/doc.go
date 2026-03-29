// Package config loads application configuration from YAML files and
// environment variables using koanf. It reads a YAML file named after
// the ENVIRONMENT variable (defaulting to "component-test.yml"), then
// overlays any matching environment variables.
//
// Infrastructure configuration sections (JWT, database, OpenTelemetry,
// server) are provided as separate fx dependencies. The *koanf.Koanf
// instance is also provided for downstream modules (e.g.
// game/logic/config) to unmarshal their own sections.
//
// # Layer
//
// Kernel — configuration loading with no internal dependencies.
package config
