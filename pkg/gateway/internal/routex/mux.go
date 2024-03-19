package routex

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/pubgo/lava/pkg/httprule"
)

// UnescapingMode defines the behavior of ServeMux when unescaping path parameters.
type UnescapingMode int

const (
	// UnescapingModeLegacy is the default V2 behavior, which escapes the entire
	// path string before doing any routing.
	UnescapingModeLegacy UnescapingMode = iota

	// UnescapingModeAllExceptReserved unescapes all path parameters except RFC 6570
	// reserved characters.
	UnescapingModeAllExceptReserved

	// UnescapingModeAllExceptSlash unescapes URL path parameters except path
	// separators, which will be left as "%2F".
	UnescapingModeAllExceptSlash

	// UnescapingModeAllCharacters unescapes all URL path parameters.
	UnescapingModeAllCharacters

	// UnescapingModeDefault is the default escaping type.
	// TODO(v3): default this to UnescapingModeAllExceptReserved per grpc-httpjson-transcoding's
	// reference implementation
	UnescapingModeDefault = UnescapingModeLegacy
)

var encodedPathSplitter = regexp.MustCompile("(/|%2F)")

// A HandlerFunc handles a specific pair of path pattern and HTTP method.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, pathParams map[string]string)

// ServeMux is a request multiplexer for grpc-gateway.
// It matches http requests to patterns and invokes the corresponding handler.
type ServeMux struct {
	// handlers maps HTTP method to a list of handlers.
	handlers       map[string][]handler
	unescapingMode UnescapingMode
}

// ServeMuxOption is an option that can be given to a ServeMux on construction.
type ServeMuxOption func(*ServeMux)

// WithUnescapingMode sets the escaping type. See the definitions of UnescapingMode
// for more information.
func WithUnescapingMode(mode UnescapingMode) ServeMuxOption {
	return func(serveMux *ServeMux) {
		serveMux.unescapingMode = mode
	}
}

// HeaderMatcherFunc checks whether a header key should be forwarded to/from gRPC context.
type HeaderMatcherFunc func(string) (string, bool)

// NewServeMux returns a new ServeMux whose internal mapping is empty.
func NewServeMux(opts ...ServeMuxOption) *ServeMux {
	serveMux := &ServeMux{
		handlers:       make(map[string][]handler),
		unescapingMode: UnescapingModeDefault,
	}

	for _, opt := range opts {
		opt(serveMux)
	}

	return serveMux
}

// Handle associates "h" to the pair of HTTP method and path pattern.
func (s *ServeMux) Handle(meth string, pat Pattern, h HandlerFunc) {
	s.handlers[meth] = append([]handler{{pat: pat, h: h}}, s.handlers[meth]...)
}

// HandlePath allows users to configure custom path handlers.
// refer: https://grpc-ecosystem.github.io/grpc-gateway/docs/operations/inject_router/
func (s *ServeMux) HandlePath(meth string, pathPattern string, h HandlerFunc) error {
	compiler, err := httprule.Parse(pathPattern)
	if err != nil {
		return fmt.Errorf("parsing path pattern: %w", err)
	}
	tp := compiler.Compile()
	pattern, err := NewPattern(tp.Version, tp.OpCodes, tp.Pool, tp.Verb)
	if err != nil {
		return fmt.Errorf("creating new pattern: %w", err)
	}
	s.Handle(meth, pattern, h)
	return nil
}

// ServeHTTP dispatches the request to the first handler whose pattern matches to r.Method and r.URL.Path.
func (s *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	path := r.URL.Path
	if !strings.HasPrefix(path, "/") {
		_, outboundMarshaler := MarshalerForRequest(s, r)
		s.routingErrorHandler(ctx, s, outboundMarshaler, w, r, http.StatusBadRequest)
		return
	}

	// TODO(v3): remove UnescapingModeLegacy
	if s.unescapingMode != UnescapingModeLegacy && r.URL.RawPath != "" {
		path = r.URL.RawPath
	}

	var pathComponents []string
	// since in UnescapeModeLegacy, the URL will already have been fully unescaped, if we also split on "%2F"
	// in this escaping mode we would be double unescaping but in UnescapingModeAllCharacters, we still do as the
	// path is the RawPath (i.e. unescaped). That does mean that the behavior of this function will change its default
	// behavior when the UnescapingModeDefault gets changed from UnescapingModeLegacy to UnescapingModeAllExceptReserved
	if s.unescapingMode == UnescapingModeAllCharacters {
		pathComponents = encodedPathSplitter.Split(path[1:], -1)
	} else {
		pathComponents = strings.Split(path[1:], "/")
	}

	lastPathComponent := pathComponents[len(pathComponents)-1]

	for _, h := range s.handlers[r.Method] {
		// If the pattern has a verb, explicitly look for a suffix in the last
		// component that matches a colon plus the verb. This allows us to
		// handle some cases that otherwise can't be correctly handled by the
		// former LastIndex case, such as when the verb literal itself contains
		// a colon. This should work for all cases that have run through the
		// parser because we know what verb we're looking for, however, there
		// are still some cases that the parser itself cannot disambiguate. See
		// the comment there if interested.

		var verb string
		patVerb := h.pat.Verb()

		idx := -1
		if patVerb != "" && strings.HasSuffix(lastPathComponent, ":"+patVerb) {
			idx = len(lastPathComponent) - len(patVerb) - 1
		}
		if idx == 0 {
			_, outboundMarshaler := MarshalerForRequest(s, r)
			s.routingErrorHandler(ctx, s, outboundMarshaler, w, r, http.StatusNotFound)
			return
		}

		comps := make([]string, len(pathComponents))
		copy(comps, pathComponents)

		if idx > 0 {
			comps[len(comps)-1], verb = lastPathComponent[:idx], lastPathComponent[idx+1:]
		}

		pathParams, err := h.pat.MatchAndEscape(comps, verb, s.unescapingMode)
		if err != nil {
			var mse MalformedSequenceError
			if ok := errors.As(err, &mse); ok {
				_, outboundMarshaler := MarshalerForRequest(s, r)
				s.errorHandler(ctx, s, outboundMarshaler, w, r, &HTTPStatusError{
					HTTPStatus: http.StatusBadRequest,
					Err:        mse,
				})
			}
			continue
		}
		h.h(w, r, pathParams)
		return
	}

	_, outboundMarshaler := MarshalerForRequest(s, r)
	s.routingErrorHandler(ctx, s, outboundMarshaler, w, r, http.StatusNotFound)
}

type handler struct {
	pat Pattern
	h   HandlerFunc
}
