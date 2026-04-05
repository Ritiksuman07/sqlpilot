package db

import (
	"fmt"
	"strings"
)

func splitQualified(name string) (string, string) {
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		return parts[0], parts[1]
	}
	return "public", name
}

func formatValue(value any) string {
	if value == nil {
		return "NULL"
	}
	switch v := value.(type) {
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}
