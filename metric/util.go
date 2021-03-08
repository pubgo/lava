package metric

import "strings"

var replacer = strings.NewReplacer(".", "_", ",", "_", " ", "_")

// StripUnsupportedCharacters cleans up a metrics key or value:
func StripUnsupportedCharacters(metricName string) string { return replacer.Replace(metricName) }
