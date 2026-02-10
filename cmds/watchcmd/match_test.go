package watchcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{
			name:     "simple include match",
			path:     "main.go",
			patterns: []string{"*.go"},
			want:     true,
		},
		{
			name:     "simple include mismatch",
			path:     "main.txt",
			patterns: []string{"*.go"},
			want:     false,
		},
		{
			name:     "path include match",
			path:     "src/main.go",
			patterns: []string{"src/*.go"},
			want:     true,
		},
		{
			name:     "doublestar include match",
			path:     "src/pkg/utils.go",
			patterns: []string{"**/*.go"},
			want:     true,
		},
		{
			name:     "exclude match",
			path:     "vendor/lib.go",
			patterns: []string{"**/*.go", "!vendor/**"},
			want:     false,
		},
		{
			name:     "exclude priority",
			path:     "main.go",
			patterns: []string{"*.go", "!main.go"},
			want:     false,
		},
		{
			name:     "filename convenience match",
			path:     "path/to/file.txt",
			patterns: []string{"*.txt"},
			want:     true,
		},
		{
			name:     "no includes means match all",
			path:     "anything.txt",
			patterns: []string{"!*.go"},
			want:     true,
		},
		{
			name:     "exclude specific file in subdirectory",
			path:     "src/ignore_me.go",
			patterns: []string{"src/*.go", "!src/ignore_me.go"},
			want:     false,
		},
		{
			name:     "multiple include match one",
			path:     "main.py",
			patterns: []string{"*.go", "*.py"},
			want:     true,
		},
		{
			name:     "multiple include match none",
			path:     "main.js",
			patterns: []string{"*.go", "*.py"},
			want:     false,
		},
		{
			name:     "complex doublestar match",
			path:     "a/b/c/d/e/f.go",
			patterns: []string{"a/**/e/*.go"},
			want:     true,
		},
		{
			name:     "complex doublestar mismatch",
			path:     "a/b/c/d/x/f.go",
			patterns: []string{"a/**/e/*.go"},
			want:     false,
		},
		{
			name:     "dotfile match",
			path:     ".gitignore",
			patterns: []string{".*"}, // .* matches dotfiles in base
			want:     true,
		},
		{
			name:     "empty patterns match everything",
			path:     "whatever.txt",
			patterns: []string{},
			want:     true,
		},
		{
			name:     "exact path match",
			path:     "cmd/main.go",
			patterns: []string{"cmd/main.go"},
			want:     true,
		},
		{
			name:     "nested exclude",
			path:     "src/vendor/pkg/file.go",
			patterns: []string{"src/**", "!**/vendor/**"},
			want:     false,
		},
		{
			name:     "windows style separator pattern",
			path:     "foo/bar.txt",
			patterns: []string{"foo\\bar.txt"}, // On unix this might be treated literally but code checks ContainsAny "/\\"
			// doublestar handle separator based on OS, but logic:
			// strings.ContainsAny(pattern, "/\\") -> true
			// doublestar.Match("foo\bar.txt", "foo/bar.txt") -> might fail on unix if it expects unix separator
			// Let's test a case that works with forward slash which is universal enough
			want: false, // Wait, if I use backslash on standard linux/mac it might match if it's treated as char, but code treats it as path separator trigger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Match(tt.path, tt.patterns)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSplitPatterns(t *testing.T) {
	patterns := []string{"*.go", "!vendor", "src/**"}
	includes, excludes := SplitPatterns(patterns)

	assert.ElementsMatch(t, []string{"*.go", "src/**"}, includes)
	assert.ElementsMatch(t, []string{"vendor"}, excludes)
}

func TestMatchAny(t *testing.T) {
	assert.True(t, MatchAny("main.go", []string{"*.go"}))
	assert.True(t, MatchAny("path/to/main.go", []string{"*.go"}))
	assert.True(t, MatchAny("path/to/main.go", []string{"path/**/*.go"}))
	assert.False(t, MatchAny("main.txt", []string{"*.go"}))
}

func TestSplitPatternsEmpty(t *testing.T) {
	includes, excludes := SplitPatterns([]string{})
	assert.Empty(t, includes)
	assert.Empty(t, excludes)
}

func TestMatchAnyEmpty(t *testing.T) {
	assert.False(t, MatchAny("anything", []string{}))
}
