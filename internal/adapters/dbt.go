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

// DbtAdapter implements the Generator interface for dbt projects.
type DbtAdapter struct{}

func NewDbtAdapter() *DbtAdapter {
	return &DbtAdapter{}
}

func (d *DbtAdapter) Name() string {
	return "dbt"
}

func (d *DbtAdapter) Generate(model core.Model, fileInfo scanner.File, dialect dialects.Dialect, outputRoot string) error {
	// 1. Create the output directory mirroring the input structure
	outDir := filepath.Join(outputRoot, fileInfo.RelativeDir)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 2. Generate SQL Content
	sqlContent, err := d.generateSQL(model, dialect)
	if err != nil {
		return fmt.Errorf("failed to generate SQL for model %s: %w", model.Name, err)
	}

	// 3. Write SQL File
	sqlPath := filepath.Join(outDir, fmt.Sprintf("%s.sql", fileInfo.BaseName))
	if err := os.WriteFile(sqlPath, []byte(sqlContent), 0644); err != nil {
		return fmt.Errorf("failed to write SQL file: %w", err)
	}

	// 4. Generate & Write schema.yml
	schemaContent := d.generateSchemaYML(model)
	schemaPath := filepath.Join(outDir, fmt.Sprintf("%s_schema.yml", fileInfo.BaseName))
	if err := os.WriteFile(schemaPath, []byte(schemaContent), 0644); err != nil {
		return fmt.Errorf("failed to write schema YAML file: %w", err)
	}

	return nil
}

func (d *DbtAdapter) generateSQL(model core.Model, dialect dialects.Dialect) (string, error) {
	var sb strings.Builder

	// Write dbt config header
	if model.Config.Materialized != "" {
		sb.WriteString("{{ config(\n")
		sb.WriteString(fmt.Sprintf("    materialized='%s'", model.Config.Materialized))

		if model.Config.UniqueKey != "" {
			sb.WriteString(fmt.Sprintf(",\n    unique_key='%s'", model.Config.UniqueKey))
		}
		if model.Config.IncrementalStrategy != "" {
			sb.WriteString(fmt.Sprintf(",\n    incremental_strategy='%s'", model.Config.IncrementalStrategy))
		}
		sb.WriteString("\n) }}\n\n")
	}

	// Build CTEs (Common Table Expressions)
	sb.WriteString("WITH base AS (\n")
	sb.WriteString(fmt.Sprintf("    %s\n", dialect.Select([]string{"*"})))

	// Handle Base Entity as a dbt source or ref
	var sourceRef string
	if model.BaseEntity.Ref != "" {
		sourceRef = fmt.Sprintf("{{ ref('%s') }}", model.BaseEntity.Ref)
	} else {
		sourceRef = fmt.Sprintf("{{ source('%s', '%s') }}", model.BaseEntity.Schema, model.BaseEntity.Table)
	}
	sb.WriteString(fmt.Sprintf("    %s\n", dialect.From(sourceRef)))
	sb.WriteString(")")

	// Build CTEs for Relationships (hiding complex JOIN setup)
	for _, relName := range sortedRelationships(model) {
		rel := model.Relationships[relName]
		sb.WriteString(fmt.Sprintf(",\n\n%s AS (\n", relName))
		sb.WriteString(fmt.Sprintf("    %s\n", dialect.Select([]string{"*"})))

		// Parse 'core.customers' into {{ ref('customers') }}
		refParts := strings.Split(rel.ToModel, ".")
		modelName := refParts[len(refParts)-1]
		sb.WriteString(fmt.Sprintf("    %s\n", dialect.From(fmt.Sprintf("{{ ref('%s') }}", modelName))))
		sb.WriteString(")")
	}

	sb.WriteString("\n\n")

	// Main SELECT Statement
	var selectCols []string
	for _, col := range model.Columns {
		// Use the globally parsed expression, or fallback to raw name
		expr := col.SourceExpression
		if expr == "" {
			expr = col.Name
		}

		// 1. Handle Relationship Pulls (e.g., customer.email)
		if col.PullFromRelationship != "" {
			expr = fmt.Sprintf("%s.%s", col.PullFromRelationship, col.TargetColumn)
		} else if col.Window != nil {
			// 2. Handle raw window functions
			expr = dialect.WindowFunction(col.Window.Function, col.Window.Column, col.Window.PartitionBy)
		} else if col.Transformation == "first_occurrence" {
			// 3. Handle Business-friendly aliases
			expr = dialect.WindowFunction("min", expr, col.PartitionBy)
		} else if !strings.Contains(expr, "(") && !strings.Contains(expr, ".") {
			// 4. Prefix base columns if no complex logic is applied
			expr = fmt.Sprintf("base.%s", expr)
		}

		selectCols = append(selectCols, fmt.Sprintf("%s AS %s", expr, col.Name))
	}

	sb.WriteString(dialect.Select(selectCols))
	sb.WriteString("\n")
	sb.WriteString(dialect.From("base"))
	sb.WriteString("\n")

	// Append JOINs for Relationships
	for _, relName := range sortedRelationships(model) {
		rel := model.Relationships[relName]
		condition := rel.MatchOnSQL
		if condition == "" {
			// Fallback to standard equi-join
			condition = fmt.Sprintf("base.%s = %s.%s", rel.MatchOn, relName, rel.MatchOn)
		}
		sb.WriteString(dialect.Join("LEFT", relName, "", condition))
		sb.WriteString("\n")
	}

	// Append UNNEST transformations (Dialect specific)
	for _, t := range model.Transformations {
		if t.Type == "unnest" {
			sb.WriteString(dialect.Unnest(t.SourceArray, t.Alias))
			sb.WriteString("\n")
		}
	}

	// Append Filters (WHERE)
	var conditions []string
	for _, filter := range model.Filters {
		op := filter.Operator
		if op == "equals" {
			op = "="
		}

		// Quote strings
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

	// Append PIVOT transformations (Dialect specific)
	for _, t := range model.Transformations {
		if t.Type == "pivot" {
			sb.WriteString(dialect.Pivot("base", t.OnColumn, t.Values, t.Aggregate))
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// generateSchemaYML automatically documents the dbt model and enforces constraints.
func (d *DbtAdapter) generateSchemaYML(model core.Model) string {
	var sb strings.Builder
	sb.WriteString("version: 2\n\n")
	sb.WriteString("models:\n")
	sb.WriteString(fmt.Sprintf("  - name: %s\n", model.Name))

	if model.Description != "" {
		sb.WriteString(fmt.Sprintf("    description: \"%s\"\n", model.Description))
	}

	if len(model.Columns) > 0 {
		sb.WriteString("    columns:\n")
		for _, col := range model.Columns {
			sb.WriteString(fmt.Sprintf("      - name: %s\n", col.Name))

			if col.Description != "" {
				sb.WriteString(fmt.Sprintf("        description: \"%s\"\n", col.Description))
			}

			hasTests := col.IsPrimaryKey || len(col.AcceptedValues) > 0
			if hasTests {
				sb.WriteString("        tests:\n")
				if col.IsPrimaryKey {
					sb.WriteString("          - unique\n          - not_null\n")
				}
				if len(col.AcceptedValues) > 0 {
					sb.WriteString("          - accepted_values:\n")
					sb.WriteString("              values: [")
					for i, val := range col.AcceptedValues {
						if i > 0 {
							sb.WriteString(", ")
						}
						sb.WriteString(fmt.Sprintf("'%s'", val))
					}
					sb.WriteString("]\n")
				}
			}
		}
	}
	return sb.String()
}
