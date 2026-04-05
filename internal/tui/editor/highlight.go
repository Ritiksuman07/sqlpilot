package editor

import (
	"bytes"

	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func highlightSQL(input string) string {
	lexer := lexers.Get("sql")
	if lexer == nil {
		return input
	}
	iterator, err := lexer.Tokenise(nil, input)
	if err != nil {
		return input
	}

	formatter := formatters.Get("terminal16m")
	if formatter == nil {
		return input
	}

	style := styles.Get("monokai")
	if style == nil {
		style = styles.Fallback
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return input
	}
	return buf.String()
}
