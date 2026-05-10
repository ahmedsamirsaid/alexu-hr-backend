package db

import (
	"fmt"
	"strings"
)

// convertPlaceholders converts ? placeholders to PostgreSQL $1, $2, $3... format
// startIndex specifies what number to start from (e.g., 1 for $1, 2 for $2, etc.)
func convertPlaceholders(query string, startIndex int) string {
	count := startIndex
	result := strings.Builder{}
	for i := 0; i < len(query); i++ {
		if query[i] == '?' {
			result.WriteString(fmt.Sprintf("$%d", count))
			count++
		} else {
			result.WriteByte(query[i])
		}
	}
	return result.String()
}

// buildPlaceholders creates PostgreSQL-style placeholders ($1, $2, $3...)
// starting from the given index
func buildPlaceholders(count int, startIndex int) string {
	if count == 0 {
		return ""
	}
	placeholders := make([]string, count)
	for i := 0; i < count; i++ {
		placeholders[i] = fmt.Sprintf("$%d", startIndex+i)
	}
	return strings.Join(placeholders, ", ")
}

// nextPlaceholder returns the next placeholder based on current args length
// e.g., if len(args) == 2, returns "$3"
func nextPlaceholder(args []any) string {
	return fmt.Sprintf("$%d", len(args)+1)
}

// nextPlaceholders returns multiple placeholders
// e.g., if len(args) == 2 and count == 2, returns "$3, $4"
func nextPlaceholders(args []any, count int) string {
	return buildPlaceholders(count, len(args)+1)
}
