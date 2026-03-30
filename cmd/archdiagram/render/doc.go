// Package render produces deterministic text outputs from the architecture model.
//
// Each renderer is a pure function: *model.ArchModel -> string.
// Renderers sort at every level to ensure deterministic output.
package render
