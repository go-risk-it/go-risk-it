// Package config provides game-specific configuration types loaded from the
// "game" section of the application YAML. It receives a *koanf.Koanf from
// the kernel config module and unmarshals the game subsection.
//
// # Layer
//
// Game-support — game-specific configuration types with no web or data imports.
package config
