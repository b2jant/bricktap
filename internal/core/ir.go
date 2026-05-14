package core

// ProjectConfig represents global project settings
type ProjectConfig struct {
	Name             string      `yaml:"name"`
	DefaultDialect   string      `yaml:"default_dialect"`
	DefaultFramework string      `yaml:"default_framework"`
	GlobalRules      GlobalRules `yaml:"global_rules"`
}

type GlobalRules struct {
	TypeCasting    map[string]string `yaml:"type_casting"`
	ColumnPatterns map[string]string `yaml:"column_patterns"`
}

// Model represents a single parsed YAML model
type Model struct {
	Name            string                  `yaml:"name"`
	Description     string                  `yaml:"description"`
	Config          ModelConfig             `yaml:"config"`
	BaseEntity      Entity                  `yaml:"base_entity"`
	Relationships   map[string]Relationship `yaml:"relationships,omitempty"`
	Transformations []Transformation        `yaml:"transformations,omitempty"`
	Columns         []Column                `yaml:"columns"`
	Filters         []Filter                `yaml:"filters,omitempty"`
}

type ModelConfig struct {
	Materialized string `yaml:"materialized"`
	Dialect      string `yaml:"dialect,omitempty"`
}

type Entity struct {
	Schema string `yaml:"schema"`
	Table  string `yaml:"table"`
}

type Relationship struct {
	ToModel    string `yaml:"to_model"`
	MatchOn    string `yaml:"match_on,omitempty"`
	MatchOnSQL string `yaml:"match_on_sql,omitempty"`
}

type Transformation struct {
	Type        string   `yaml:"type"`
	SourceArray string   `yaml:"source_array,omitempty"`
	Alias       string   `yaml:"alias,omitempty"`
	OnColumn    string   `yaml:"on_column,omitempty"`
	Values      []string `yaml:"values,omitempty"`
	Aggregate   string   `yaml:"aggregate,omitempty"`
}

type Column struct {
	Name                 string        `yaml:"name"`
	Description          string        `yaml:"description,omitempty"`
	SourceColumn         string        `yaml:"source_column,omitempty"`
	Type                 string        `yaml:"type,omitempty"`
	IsPrimaryKey         bool          `yaml:"is_primary_key,omitempty"`
	PullFromRelationship string        `yaml:"pull_from_relationship,omitempty"`
	TargetColumn         string        `yaml:"target_column,omitempty"`
	Transformation       string        `yaml:"transformation,omitempty"`
	PartitionBy          []string      `yaml:"partition_by,omitempty"`
	Window               *WindowConfig `yaml:"window,omitempty"`
	SourceExpression     string        // Injected during parsing based on global rules
}

type WindowConfig struct {
	Function    string   `yaml:"function"`
	Column      string   `yaml:"column"`
	PartitionBy []string `yaml:"partition_by"`
}

type Filter struct {
	Column   string `yaml:"column"`
	Operator string `yaml:"operator"`
	Value    string `yaml:"value"`
}
