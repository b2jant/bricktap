package dialects

import (
	"fmt"
	"strings"
)

// SnowflakeDialect implements Snowflake-specific SQL generation.
type SnowflakeDialect struct {
	BaseDialect
}

func NewSnowflakeDialect() *SnowflakeDialect {
	return &SnowflakeDialect{}
}

func (s *SnowflakeDialect) Name() string {
	return "snowflake"
}

// Unnest in Snowflake uses LATERAL FLATTEN
func (s *SnowflakeDialect) Unnest(arrayCol string, alias string) string {
	// e.g. , LATERAL FLATTEN(input => properties.events) event_data
	return fmt.Sprintf(", LATERAL FLATTEN(input => %s) %s", arrayCol, alias)
}

// Pivot in Snowflake uses the PIVOT keyword
func (s *SnowflakeDialect) Pivot(source string, onCol string, values []string, agg string) string {
	// e.g. PIVOT(COUNT(event_name) FOR event_name IN ('login', 'purchase'))
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}

	return fmt.Sprintf("PIVOT(%s(%s) FOR %s IN (%s))",
		strings.ToUpper(agg),
		onCol,
		onCol,
		strings.Join(quotedValues, ", "))
}
