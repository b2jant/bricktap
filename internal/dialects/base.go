package dialects

import (
	"fmt"
	"strings"
)

// BaseDialect provides common ANSI SQL implementations that most warehouses share.
type BaseDialect struct{}

func (b *BaseDialect) Name() string {
	return "ansi"
}

func (b *BaseDialect) QuoteIdentifier(identifier string) string {
	// Standard ANSI quoting, though many warehouses prefer double quotes
	return fmt.Sprintf(`"%s"`, identifier)
}

func (b *BaseDialect) Select(columns []string) string {
	if len(columns) == 0 {
		return "SELECT *"
	}
	return "SELECT\n    " + strings.Join(columns, ",\n    ")
}

func (b *BaseDialect) From(table string) string {
	return fmt.Sprintf("FROM %s", table)
}

func (b *BaseDialect) Join(joinType string, table string, alias string, condition string) string {
	if alias != "" {
		table = fmt.Sprintf("%s AS %s", table, alias)
	}
	return fmt.Sprintf("%s JOIN %s\n  ON %s", strings.ToUpper(joinType), table, condition)
}

func (b *BaseDialect) Where(conditions []string) string {
	if len(conditions) == 0 {
		return ""
	}
	return "WHERE " + strings.Join(conditions, "\n  AND ")
}

func (b *BaseDialect) WindowFunction(funcName string, col string, partition []string) string {
	partitionClause := ""
	if len(partition) > 0 {
		partitionClause = fmt.Sprintf("PARTITION BY %s", strings.Join(partition, ", "))
	}

	// Base ANSI window function
	return fmt.Sprintf("%s(%s) OVER (%s)", strings.ToUpper(funcName), col, partitionClause)
}

func (b *BaseDialect) Unnest(arrayCol string, alias string) string {
	return "-- WARNING: Unnest is not supported in the Base ANSI dialect. Please specify a target data warehouse."
}

func (b *BaseDialect) Pivot(source string, onCol string, values []string, agg string) string {
	return "-- WARNING: Pivot is not supported in the Base ANSI dialect. Please specify a target data warehouse."
}
