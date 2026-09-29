// internal/types/types.go
package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

// TextArray wraps []string for PostgreSQL text[] type
type TextArray []string

// Value implements driver.Valuer interface
func (ta TextArray) Value() (driver.Value, error) {
	if len(ta) == 0 {
		return "{}", nil
	}
	// Escape and wrap in quotes
	escaped := make([]string, len(ta))
	for i, v := range ta {
		// Escape double quotes
		escaped[i] = `"` + strings.ReplaceAll(v, `"`, `""`) + `"`
	}
	return "{" + strings.Join(escaped, ",") + "}", nil
}

// Scan implements sql.Scanner interface
func (ta *TextArray) Scan(value interface{}) error {
	if value == nil {
		*ta = TextArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*ta = parseArray(string(v))
	case string:
		*ta = parseArray(v)
	default:
		return errors.New(fmt.Sprintf("cannot scan %T into TextArray", value))
	}
	return nil
}

// parseArray parses PostgreSQL array string like {a,b,"c,d"} into []string
func parseArray(s string) TextArray {
	if s == "" || s == "{}" {
		return TextArray{}
	}
	if len(s) < 2 || s[0] != '{' || s[len(s)-1] != '}' {
		return TextArray{s}
	}

	s = s[1 : len(s)-1]
	if s == "" {
		return TextArray{}
	}

	var result TextArray
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(s); i++ {
		r := rune(s[i])

		switch r {
		case '"':
			if inQuote && i > 0 && s[i-1] == '\\' {
				current.WriteRune(r)
			} else {
				inQuote = !inQuote
			}
		case ',':
			if inQuote {
				current.WriteRune(r)
			} else {
				result = append(result, strings.Trim(current.String(), `\"`))
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		result = append(result, strings.Trim(current.String(), `\"`))
	}

	return result
}
