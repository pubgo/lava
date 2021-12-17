package internal

import (
	"strings"
)

func unExport(s string) string { return strings.ToLower(s[:1]) + s[1:] }

const deprecationComment = "// Deprecated: Do not use."
