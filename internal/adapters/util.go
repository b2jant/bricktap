package adapters

import (
	"sort"

	"github.com/b2jant/bricktap/internal/core"
)

// sortedRelationships returns the keys of the relationships map in deterministic order.
func sortedRelationships(model core.Model) []string {
	var keys []string
	for k := range model.Relationships {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
