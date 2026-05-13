package dialects

// Dialect defines how specific SQL operations are written for different databases.
type Dialect interface {
	Name() string
	QuoteIdentifier(identifier string) string

	// Base SQL operations
	Select(columns []string) string
	From(table string) string
	Join(joinType string, table string, alias string, condition string) string
	Where(conditions []string) string

	// Advanced Transformations
	Unnest(arrayCol string, alias string) string
	Pivot(source string, onCol string, values []string, agg string) string
	WindowFunction(funcName string, col string, partition []string) string
}
