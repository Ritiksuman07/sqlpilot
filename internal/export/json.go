package export

import (
	"encoding/json"
	"io"
)

func WriteJSON(w io.Writer, columns []string, rows [][]string) error {
	objects := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		obj := map[string]string{}
		for i, col := range columns {
			if i < len(row) {
				obj[col] = row[i]
			}
		}
		objects = append(objects, obj)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(objects)
}
