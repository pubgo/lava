package prometheus

import "strings"

var replacer = strings.NewReplacer(".", "_", ",", "_", " ", "_", "__", "_")

// StripUnsupportedCharacters cleans up a metrics key or value:
func StripUnsupportedCharacters(metricName string) string { return replacer.Replace(metricName) }
