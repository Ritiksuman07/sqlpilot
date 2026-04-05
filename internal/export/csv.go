package export

import (
	"encoding/csv"
	"io"
)

func WriteCSV(w io.Writer, columns []string, rows [][]string) error {
	writer := csv.NewWriter(w)
	if len(columns) > 0 {
		if err := writer.Write(columns); err != nil {
			return err
		}
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
