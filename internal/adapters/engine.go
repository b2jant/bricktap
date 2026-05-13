package adapters

import (
	"github.com/b2jant/bricktap/internal/core"
	"github.com/b2jant/bricktap/internal/dialects"
	"github.com/b2jant/bricktap/internal/scanner"
)

// Generator is the interface that Framework Adapters (dbt, sqlmesh, etc.) must implement.
type Generator interface {
	Name() string
	Generate(model core.Model, fileInfo scanner.File, dialect dialects.Dialect, outputRoot string) error
}
