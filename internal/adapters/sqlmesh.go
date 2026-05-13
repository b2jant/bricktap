package adapters

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/b2jant/bricktap/internal/core"
	"github.com/b2jant/bricktap/internal/dialects"
	"github.com/b2jant/bricktap/internal/scanner"
)

// SqlMeshAdapter implements the Generator interface for SQLMesh projects.
type SqlMeshAdapter struct{}

func NewSqlMeshAdapter() *SqlMeshAdapter {
	return &SqlMeshAdapter{}
}

func (s *SqlMeshAdapter) Name() string {
	return "sqlmesh"
}

func (s *SqlMeshAdapter) Generate(model core.Model, fileInfo scanner.File, dialect dialects.Dialect, outputRoot string) error {
	// 1. Create the output directory mirroring the input structure
	outDir := filepath.Join(outputRoot, fileInfo.RelativeDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 2. Generate SQL Content
	sqlContent, err := s.generateSQL(model, dialect)
	if err != nil {
		return fmt.Errorf("failed to generate SQL for model %s: %w", model.Name, err)
	}

	// 3. Write SQL File
	sqlPath := filepath.Join(outDir, fmt.Sprintf("%s.sql", fileInfo.BaseName))
	if err := os.WriteFile(sqlPath, []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("failed to write SQL file: %w", err)
	}

	return nil
}

func (s *SqlMeshAdapter) generateSQL(model core.Model, dialect dialects.Dialect) (string, error) {
	var sb strings.Builder

	// Write SQLMesh MODEL header
	sb.WriteString("MODEL (\n")
	sb.WriteString(fmt.Sprintf("  name %s,\n", model.Name)) // Often schema.name in SQLMesh

	kind := "VIEW"
	if model.Config.Materialized == "table" {
		kind = "FULL"
	}
	sb.WriteString(fmt.Sprintf("  kind %s", kind))

	if model.Description != "" {
		sb.WriteString(fmt.Sprintf(",\n  description '%s'\n", model.Description))
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString(");\n\n")

	// Build CTEs
	sb.WriteString("WITH base AS (\n")
	sb.WriteString(fmt.Sprintf("    %s\n", dialect.Select([]string{"*"})))

	// Handle Base Entity as a source table
	sourceRef := fmt.Sprintf("%s.%s", model.BaseEntity.Schema, model.BaseEntity.Table)
	sb.WriteString(fmt.Sprintf("    %s\n", dialect.From(sourceRef)))
	sb.WriteString(")")

	// Build CTEs for Relationships
	for relName, rel := range model.Relationships {
		sb.WriteString(fmt.Sprintf(",\n\n%s AS (\n", relName))
		sb.WriteString(fmt.Sprintf("    %s\n", dialect.Select([]string{"*"})))
		sb.WriteString(fmt.Sprintf("    %s\n", dialect.From(rel.ToModel)))
		sb.WriteString(")")
	}

	sb.WriteString("\n\n")

	// Main SELECT Statement
	var selectCols []string
	for _, col := range model.Columns {
		expr := col.SourceExpression
		if expr == "" {
			expr = col.Name
		}

		if col.PullFromRelationship != "" {
			expr = fmt.Sprintf("%s.%s", col.PullFromRelationship, col.TargetColumn)
		} else if col.Window != nil {
			expr = dialect.WindowFunction(col.Window.Function, col.Window.Column, col.Window.PartitionBy)
		} else if col.Transformation == "first_occurrence" {
			expr = dialect.WindowFunction("min", expr, col.PartitionBy)
		} else if !strings.Contains(expr, "(") && !strings.Contains(expr, ".") {
			expr = fmt.Sprintf("base.%s", expr)
		}

		selectCols = append(selectCols, fmt.Sprintf("%s AS %s", expr, col.Name))
	}

	sb.WriteString(dialect.Select(selectCols))
	sb.WriteString("\n")
	sb.WriteString(dialect.From("base"))
	sb.WriteString("\n")

	for relName, rel := range model.Relationships {
		condition := rel.MatchOnSQL
		if condition == "" {
			condition = fmt.Sprintf("base.%s = %s.%s", rel.MatchOn, relName, rel.MatchOn)
		}
		sb.WriteString(dialect.Join("LEFT", relName, "", condition))
		sb.WriteString("\n")
	}

	for _, t := range model.Transformations {
		if t.Type == "unnest" {
			sb.WriteString(dialect.Unnest(t.SourceArray, t.Alias))
			sb.WriteString("\n")
		}
	}

	var conditions []string
	for _, filter := range model.Filters {
		op := filter.Operator
		if op == "equals" {
			op = "="
		}

		val := filter.Value
		if val != "true" && val != "false" && !strings.HasPrefix(val, "'") {
			val = fmt.Sprintf("'%s'", val)
		}
		conditions = append(conditions, fmt.Sprintf("base.%s %s %s", filter.Column, op, val))
	}

	if len(conditions) > 0 {
		sb.WriteString(dialect.Where(conditions))
		sb.WriteString("\n")
	}

	for _, t := range model.Transformations {
		if t.Type == "pivot" {
			sb.WriteString(dialect.Pivot("base", t.OnColumn, t.Values, t.Aggregate))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}
