package watchcmd

import (
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// Match checks if the path matches the patterns logic.
// Patterns can include inclusion patterns and exclusion patterns (starting with !).
// Exclusions take precedence.
// If no inclusion patterns are present, it implies matching everything (unless excluded).
func Match(path string, patterns []string) bool {
	includes, excludes := SplitPatterns(patterns)

	// 1. Check excludes first
	if MatchAny(path, excludes) {
		return false
	}

	// 2. Check includes
	// If no include patterns are provided, we assume everything matches (subject to exclusions)
	if len(includes) == 0 {
		return true
	}

	return MatchAny(path, includes)
}

// SplitPatterns splits patterns into includes and excludes.
// Patterns starting with '!' are considered excludes (with '!' stripped).
func SplitPatterns(patterns []string) (includes, excludes []string) {
	for _, p := range patterns {
		if strings.HasPrefix(p, "!") {
			excludes = append(excludes, strings.TrimPrefix(p, "!"))
		} else {
			includes = append(includes, p)
		}
	}
	return
}

// MatchAny checks if path matches any of the provided patterns.
// It supports doublestar syntax.
// If a pattern contains path separators, it matches against the full path.
// Otherwise, it matches against the base filename.
func MatchAny(path string, patterns []string) bool {
	for _, pattern := range patterns {
		// If pattern contains path separator, match against full path
		if strings.ContainsAny(pattern, "/\\") {
			if matched, _ := doublestar.Match(pattern, path); matched {
				return true
			}
		} else {
			// Otherwise match against filename only
			if matched, _ := doublestar.Match(pattern, filepath.Base(path)); matched {
				return true
			}
		}
	}
	return false
}
