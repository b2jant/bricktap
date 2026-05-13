package parser

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/b2jant/bricktap/internal/core"
)

// ParseModel reads a YAML file and unmarshals it into the IR Model.
func ParseModel(filePath string, rules core.GlobalRules) (*core.Model, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	var model core.Model
	if err := yaml.Unmarshal(data, &model); err != nil {
		return nil, fmt.Errorf("failed to parse YAML in %s: %w", filePath, err)
	}

	// Default fallback for name if not specified
	if model.Name == "" {
		// This should be handled higher up using the filename, but we ensure it's not empty
		model.Name = "unnamed_model"
	}

	// Apply global rules to the parsed columns
	for i, col := range model.Columns {
		model.Columns[i].SourceExpression = applyRules(col, rules)
	}

	return &model, nil
}

// applyRules computes the SQL expression for a column based on global typing and pattern rules.
func applyRules(col core.Column, rules core.GlobalRules) string {
	// Start with the raw column name (or an explicit source column)
	expr := col.Name
	if col.SourceColumn != "" {
		expr = col.SourceColumn
	}

	// 1. Apply regex pattern rules first (e.g. ^is_.*$)
	for pattern, replacement := range rules.ColumnPatterns {
		matched, _ := regexp.MatchString(pattern, expr)
		if matched {
			expr = strings.ReplaceAll(replacement, "{column}", expr)
		}
	}

	// 2. Apply type-casting rules (e.g. string -> NULLIF(TRIM(x), ''))
	if col.Type != "" {
		if castExpr, exists := rules.TypeCasting[col.Type]; exists {
			expr = strings.ReplaceAll(castExpr, "{column}", expr)
		}
	}

	return expr
}
