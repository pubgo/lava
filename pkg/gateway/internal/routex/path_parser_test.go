// Copyright 2023-2024 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package routex

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPath_ParsePathTemplate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		tmpl        string
		wantPath    []string
		wantVerb    string
		wantVars    []pathVariable
		expectedErr string
	}{{
		// no error, but lots of encoding for special/reserved characters
		tmpl:     "/my%2fcool-blog~about%2Cstuff%5Bwat%5D/{var={abc=%2F%2A}/{def=%2F%2A/**}}:baz",
		wantPath: []string{"my%2Fcool-blog~about%2Cstuff%5Bwat%5D", "%2F%2A", "%2F%2A", "**"},
		wantVerb: "baz",
		wantVars: []pathVariable{
			{FieldPath: "var", start: 1, end: -1},
			{FieldPath: "abc", start: 1, end: 2},
			{FieldPath: "def", start: 2, end: -1},
		},
	}, {
		tmpl:        "/{1}",
		expectedErr: "syntax error at column 3: expected identifier",
	}, {
		tmpl:        "/{field.1}",
		expectedErr: "syntax error at column 9: expected identifier",
	}, {
		tmpl:     "/{_}",
		wantPath: []string{"*"},
		wantVars: []pathVariable{
			{FieldPath: "_", start: 0, end: 1},
		},
	}, {
		tmpl:        "/{-}",
		expectedErr: "syntax error at column 3: expected identifier",
	}, {
		tmpl:        "/{field-}",
		expectedErr: "syntax error at column 8: expected '}', got '-'",
	}, {
		tmpl:        "/foo/bar/baz?abc=def",
		expectedErr: "syntax error at column 13: unexpected '?'", // no query string allowed
	}, {
		tmpl:        "/foo/bar/baz buzz",
		expectedErr: "syntax error at column 13: unexpected ' '", // no whitespace allowed
	}, {
		tmpl:        "foo/bar/baz",
		expectedErr: "syntax error at column 1: expected '/', got 'f'", // must start with slash
	}, {
		tmpl:        "/foo/bar/",
		expectedErr: "syntax error at column 9: expected HttpPath Value", // must not end in slash
	}, {
		tmpl:        "/foo/bar:baz/buzz",
		expectedErr: "syntax error at column 13: unexpected '/'", // ":baz" Verb can only come at the very end
	}, {
		tmpl:        "/foo/{bar/baz}/buzz",
		expectedErr: "syntax error at column 10: expected '}', got '/'", // invalid field HttpPath
	}, {
		tmpl:     "/foo/bar:baz%12xyz%abcde",
		wantPath: []string{"foo", "bar"},
		wantVerb: "baz%12xyz%ABcde",
	}, {
		tmpl:     "/{hello}/world",
		wantPath: []string{"*", "world"},
		wantVars: []pathVariable{
			{FieldPath: "hello", start: 0, end: 1},
		},
	}, {
		tmpl:        "/foo/bar%55:baz%1",
		expectedErr: "syntax error at column 17: invalid URL escape \"%1\"",
	}, {
		tmpl:        "/foo/bar*",
		expectedErr: "syntax error at column 9: unexpected '*'", // wildcard must be entire HttpPath component
	}, {
		tmpl:        "/foo/bar/***",
		expectedErr: "syntax error at column 12: unexpected '*'", // no such thing as triple-wildcard
	}, {
		tmpl:        "/foo/**/bar",
		expectedErr: "double wildcard '**' must be the final HttpPath segment", // double-wildcard must be at end
	}, {
		tmpl:        "/{a}/{a}", // TODO: allow this?
		expectedErr: "duplicate variable \"a\"",
	}, {
		tmpl:     "/f/bar",
		wantPath: []string{"f", "bar"},
	}, {
		tmpl:     "/v1/{Name=shelves/*/books/*}",
		wantPath: []string{"v1", "shelves", "*", "books", "*"},
		wantVars: []pathVariable{
			{FieldPath: "Name", start: 1, end: 5},
		},
	}, {
		tmpl:     "/v1/{parent=shelves/*}/books",
		wantPath: []string{"v1", "shelves", "*", "books"},
		wantVars: []pathVariable{
			{FieldPath: "parent", start: 1, end: 3},
		},
	}, {
		tmpl:     "/v1/{book.Name=shelves/*/books/*}",
		wantPath: []string{"v1", "shelves", "*", "books", "*"},
		wantVars: []pathVariable{
			{FieldPath: "book.Name", start: 1, end: 5},
		},
	}, {
		tmpl:     "/v1:watch",
		wantPath: []string{"v1"},
		wantVerb: "watch",
	}, {
		tmpl:     "/v3/events:clear",
		wantPath: []string{"v3", "events"},
		wantVerb: "clear",
	}, {
		tmpl:     "/v3/{Name=events/*}:cancel",
		wantPath: []string{"v3", "events", "*"},
		wantVerb: "cancel",
		wantVars: []pathVariable{
			{FieldPath: "Name", start: 1, end: 3},
		},
	}, {
		tmpl:     "/foo/bar/baz/buzz",
		wantPath: []string{"foo", "bar", "baz", "buzz"},
	}, {
		tmpl:     "/foo/bar/{Name}",
		wantPath: []string{"foo", "bar", "*"},
		wantVars: []pathVariable{
			{FieldPath: "Name", start: 2, end: 3},
		},
	}, {
		tmpl:     "/foo/bar/{Name}/baz/{child}",
		wantPath: []string{"foo", "bar", "*", "baz", "*"},
		wantVars: []pathVariable{
			{FieldPath: "Name", start: 2, end: 3},
			{FieldPath: "child", start: 4, end: 5},
		},
	}, {
		tmpl:     "/foo/bar/{Name}/baz/{child.id}/buzz/{child.thing.id}",
		wantPath: []string{"foo", "bar", "*", "baz", "*", "buzz", "*"},
		wantVars: []pathVariable{
			{FieldPath: "Name", start: 2, end: 3},
			{FieldPath: "child.id", start: 4, end: 5},
			{FieldPath: "child.thing.id", start: 6, end: 7},
		},
	}, {
		tmpl:     "/foo/bar/*/{thing.id}/{cat=**}",
		wantPath: []string{"foo", "bar", "*", "*", "**"},
		wantVars: []pathVariable{
			{FieldPath: "thing.id", start: 3, end: 4},
			{FieldPath: "cat", start: 4, end: -1},
		},
	}, {
		tmpl:     "/foo/bar/*/{thing.id}/{cat=**}:do",
		wantPath: []string{"foo", "bar", "*", "*", "**"},
		wantVerb: "do",
		wantVars: []pathVariable{
			{FieldPath: "thing.id", start: 3, end: 4},
			{FieldPath: "cat", start: 4, end: -1},
		},
	}, {
		tmpl:     "/foo/bar/*/{thing.id}/{cat=**}:cancel",
		wantPath: []string{"foo", "bar", "*", "*", "**"},
		wantVerb: "cancel",
		wantVars: []pathVariable{
			{FieldPath: "thing.id", start: 3, end: 4},
			{FieldPath: "cat", start: 4, end: -1},
		},
	}, {
		tmpl:     "/foo/bob/{book_id={author}/{isbn}/*}/details",
		wantPath: []string{"foo", "bob", "*", "*", "*", "details"},
		wantVars: []pathVariable{
			{FieldPath: "book_id", start: 2, end: 5},
			{FieldPath: "author", start: 2, end: 3},
			{FieldPath: "isbn", start: 3, end: 4},
		},
	}, {
		tmpl: "/foo/blah/{longest_var={long_var.a={medium.a={short.aa}/*/{short.ab}/foo}/*}/{long_var.b={medium.b={short.ba}/*/{short.bb}/foo}/{last=**}}}:details",
		wantPath: []string{
			"foo", "blah",
			"*",   // 2 logest_var, long_var.a, medium.a, short.aa
			"*",   // 3
			"*",   // 4 short.ab
			"foo", // 5
			"*",   // 6
			"*",   // 7 long_var.b, medium.b, short.ba
			"*",   // 8
			"*",   // 9 short.bb
			"foo", // 10
			"**",  // 11 last
		},
		wantVerb: "details",
		wantVars: []pathVariable{
			{FieldPath: "longest_var", start: 2, end: -1},
			{FieldPath: "long_var.a", start: 2, end: 7},
			{FieldPath: "medium.a", start: 2, end: 6},
			{FieldPath: "short.aa", start: 2, end: 3},
			{FieldPath: "short.ab", start: 4, end: 5},
			{FieldPath: "long_var.b", start: 7, end: -1},
			{FieldPath: "medium.b", start: 7, end: 11},
			{FieldPath: "short.ba", start: 7, end: 8},
			{FieldPath: "short.bb", start: 9, end: 10},
			{FieldPath: "last", start: 11, end: -1},
		},
	}, {
		tmpl:     "/foo%2Fbar/%2A/%2A%2a/{starstar=%2A%2a/**}:%2c",
		wantPath: []string{"foo%2Fbar", "%2A", "%2A%2A", "%2A%2A", "**"},
		wantVerb: "%2C",
		wantVars: []pathVariable{
			{FieldPath: "starstar", start: 3, end: -1},
		},
	}}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.tmpl, func(t *testing.T) {
			t.Parallel()
			segments, variables, err := parsePathTemplate(testCase.tmpl)
			if testCase.expectedErr != "" {
				assert.ErrorContains(t, err, testCase.expectedErr)
				return
			}
			t.Log(segments)
			require.NoError(t, err)
			assert.ElementsMatch(t, testCase.wantPath, segments.path, "HttpPath mismatch")
			assert.Equal(t, testCase.wantVerb, segments.verb, "Verb mismatch")
			assert.ElementsMatch(t, testCase.wantVars, variables, "variables mismatch")
		})
	}
}

func TestPath_SafeLiterals(t *testing.T) {
	t.Parallel()
	literalvalues := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._%25~"
	for _, r := range literalvalues {
		if !isLiteral(r) {
			t.Errorf("isLiteral(%q) = false, want true", r)
		}
	}
	unescaped, err := url.PathUnescape(literalvalues)
	require.NoError(t, err)
	escaped := url.PathEscape(unescaped)
	assert.Equal(t, literalvalues, escaped)
}

func TestPath_Escaping(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		input       string
		mode        pathEncoding
		want        string
		wantEscaped string
		wantErr     string
	}{{
		input: "foo",
		mode:  pathEncodeSingle,
		want:  "foo",
	}, {
		input: "foo%2Fbar",
		mode:  pathEncodeSingle,
		want:  "foo/bar",
	}, {
		input: "foo%252Fbar",
		mode:  pathEncodeSingle,
		want:  "foo%2Fbar",
	}, {
		input: "foo%2Fbar",
		mode:  pathEncodeMulti,
		want:  "foo%2Fbar",
	}, {
		input:       "foo%2fbar",
		mode:        pathEncodeMulti,
		want:        "foo%2Fbar",
		wantEscaped: "foo%2Fbar",
	}}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()
			dec, err := pathUnescape(testCase.input, testCase.mode)
			if err != nil {
				assert.EqualError(t, err, testCase.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, testCase.want, dec)
			enc := pathEscape(dec, testCase.mode)
			if testCase.wantEscaped != "" {
				assert.Equal(t, testCase.wantEscaped, enc)
			} else {
				assert.Equal(t, testCase.input, enc)
			}
		})
	}
}
