package curlcmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/errors"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/pkg/cliutil"
)

type kvFlag struct {
	data map[string]string
}

func (k *kvFlag) Set(v string) error {
	parts := strings.SplitN(v, "=", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid key=value pair: %s", v)
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])
	if key == "" {
		return fmt.Errorf("invalid key in pair: %s", v)
	}

	if k.data == nil {
		k.data = make(map[string]string)
	}
	k.data[key] = val
	return nil
}

func (k *kvFlag) String() string {
	if k == nil {
		return ""
	}

	keys := make([]string, 0, len(k.data))
	for key := range k.data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, k.data[key]))
	}
	return strings.Join(parts, ",")
}

func (k *kvFlag) Map() map[string]string {
	if k == nil {
		return nil
	}
	return k.data
}

func (k *kvFlag) Type() string { return "kv" }

type gatewayVarInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type routeOperation struct {
	Method    string         `json:"method"`
	Path      string         `json:"path"`
	Operation string         `json:"operation"`
	Verb      string         `json:"verb"`
	Vars      []string       `json:"vars"`
	Extras    map[string]any `json:"extras"`
}

type gatewayInfo struct {
	Method []routeOperation `json:"method"`
}

// New returns curl command which provides a lightweight HTTP client for gateway APIs.
func New() *redant.Command {
	defaultAddr := fmt.Sprintf("http://127.0.0.1:%d", running.HttpPort.Value())

	var (
		addr           = defaultAddr
		prefix         = "/api"
		opFlag         string
		pathFlag       string
		methodOverride string
		data           string
		dataFile       string
		readStdin      bool
		listOnly       bool
		varName        = "grpc-server-info"
		timeout        = 15 * time.Second
		insecure       bool
		noPretty       bool

		headers = &kvFlag{}
		queries = &kvFlag{}
		params  = &kvFlag{}
	)

	cmd := &redant.Command{
		Use:   "curl [flags] <operation|path>",
		Short: cliutil.UsageDesc("%s gateway curl helper", version.Project()),
		Options: redant.OptionSet{
			{Flag: "addr", Description: "gateway http address, e.g. http://127.0.0.1:8080", Default: addr, Value: redant.StringOf(&addr)},
			{Flag: "prefix", Description: "gateway prefix path", Default: prefix, Value: redant.StringOf(&prefix)},
			{Flag: "operation", Description: "operation name (grpc full method or custom name)", Value: redant.StringOf(&opFlag)},
			{Flag: "path", Description: "explicit request path, overrides operation", Value: redant.StringOf(&pathFlag)},
			{Flag: "method", Shorthand: "X", Description: "http method override", Value: redant.StringOf(&methodOverride)},
			{Flag: "data", Shorthand: "d", Description: "request body string", Value: redant.StringOf(&data)},
			{Flag: "data-file", Description: "path to file used as request body", Value: redant.StringOf(&dataFile)},
			{Flag: "stdin", Description: "read request body from stdin", Value: redant.BoolOf(&readStdin)},
			{Flag: "list", Description: "list registered gateway routes and exit", Value: redant.BoolOf(&listOnly)},
			{Flag: "vars-name", Description: "expvar name that stores gateway info", Default: varName, Value: redant.StringOf(&varName)},
			{Flag: "timeout", Description: "request timeout, e.g. 5s or 2m", Default: timeout.String(), Value: redant.DurationOf(&timeout)},
			{Flag: "insecure", Shorthand: "k", Description: "allow insecure TLS connections", Value: redant.BoolOf(&insecure)},
			{Flag: "raw", Description: "print response body as-is without JSON pretty format", Value: redant.BoolOf(&noPretty)},
			{Flag: "header", Shorthand: "H", Description: "set request header, key=value (repeatable)", Value: headers},
			{Flag: "query", Shorthand: "Q", Description: "set query parameter, key=value (repeatable)", Value: queries},
			{Flag: "param", Shorthand: "P", Description: "set path parameter, key=value (repeatable)", Value: params},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			remaining := inv.Args
			target := ""
			if len(remaining) > 0 {
				target = remaining[0]
			}
			if pathFlag != "" {
				target = pathFlag
			}
			if opFlag != "" {
				target = opFlag
			}

			client := &http.Client{Timeout: timeout}
			if insecure {
				tr := http.DefaultTransport.(*http.Transport).Clone()
				tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
				client.Transport = tr
			}

			ctxReq, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			routes := make([]routeOperation, 0)
			needRoutes := listOnly || (target != "" && !strings.HasPrefix(target, "/"))
			if needRoutes {
				rts, err := fetchGatewayRoutes(ctxReq, client, addr, varName)
				if err != nil {
					return err
				}
				routes = rts
			}

			if listOnly {
				if err := printRoutes(os.Stdout, routes, prefix); err != nil {
					return errors.Wrap(err, "failed to print routes")
				}
				return nil
			}

			if target == "" {
				return errors.New("operation or path is required")
			}

			var matched *routeOperation
			if !strings.HasPrefix(target, "/") {
				for idx := range routes {
					if routes[idx].Operation == target {
						matched = &routes[idx]
						break
					}
				}
				if matched == nil {
					return errors.Errorf("operation %q not found, use --list to inspect", target)
				}
			}

			method := methodOverride
			rawPath := target
			if matched != nil {
				method = normalizeMethod(matched.Method)
				rawPath = matched.Path
			}
			if method == "" {
				method = http.MethodGet
			}

			fullPath := joinPath(prefix, rawPath)
			filledPath, err := applyPathParams(fullPath, params.Map())
			if err != nil {
				return err
			}
			fullPath = filledPath

			fullURL := strings.TrimRight(addr, "/") + fullPath
			reqBody, err := buildRequestBody(data, dataFile, readStdin)
			if err != nil {
				return err
			}

			req, err := http.NewRequestWithContext(ctxReq, method, fullURL, reqBody)
			if err != nil {
				return errors.Wrap(err, "failed to build request")
			}

			if q := queries.Map(); len(q) > 0 {
				query := req.URL.Query()
				for k, v := range q {
					query.Set(k, v)
				}
				req.URL.RawQuery = query.Encode()
			}

			if h := headers.Map(); len(h) > 0 {
				for k, v := range h {
					req.Header.Set(k, v)
				}
			}

			// auto inject token if present and Authorization not set
			if req.Header.Get("Authorization") == "" {
				if token, _ := loadToken(); token != "" {
					req.Header.Set("Authorization", "Bearer "+token)
				}
			}

			if req.Body != nil && req.Header.Get("Content-Type") == "" {
				req.Header.Set("Content-Type", "application/json")
			}
			if req.Header.Get("Accept") == "" {
				req.Header.Set("Accept", "application/json")
			}

			start := time.Now()
			resp, err := client.Do(req)
			if err != nil {
				return errors.Wrap(err, "request failed")
			}

			elapsed := time.Since(start)
			bodyBytes, err := io.ReadAll(resp.Body)
			closeErr := resp.Body.Close()
			if err != nil {
				return errors.Wrap(err, "failed to read response body")
			}
			if closeErr != nil {
				return errors.Wrap(closeErr, "failed to close response body")
			}

			if _, err := fmt.Fprintf(os.Stdout, "=> %s %s\n", method, req.URL.String()); err != nil {
				return errors.Wrap(err, "failed to write request line")
			}
			if _, err := fmt.Fprintf(os.Stdout, "<= %d %s (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode), elapsed); err != nil {
				return errors.Wrap(err, "failed to write response line")
			}
			if err := printHeaders(os.Stdout, resp.Header); err != nil {
				return errors.Wrap(err, "failed to write headers")
			}
			if len(bodyBytes) > 0 {
				if _, err := fmt.Fprintln(os.Stdout); err != nil {
					return errors.Wrap(err, "failed to write newline")
				}
				if !noPretty && strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "json") {
					var buf bytes.Buffer
					if err := json.Indent(&buf, bodyBytes, "", "  "); err == nil {
						_, _ = buf.WriteTo(os.Stdout)
						if buf.Len() == 0 || bodyBytes[len(bodyBytes)-1] != '\n' {
							if _, err := fmt.Fprintln(os.Stdout); err != nil {
								return errors.Wrap(err, "failed to write newline")
							}
						}
						return nil
					}
				}
				if _, err := os.Stdout.Write(bodyBytes); err != nil {
					return errors.Wrap(err, "failed to write body")
				}
				if bodyBytes[len(bodyBytes)-1] != '\n' {
					if _, err := fmt.Fprintln(os.Stdout); err != nil {
						return errors.Wrap(err, "failed to write newline")
					}
				}
			}

			return nil
		},
	}

	cmd.Children = append(cmd.Children, newLoginCommand())

	return cmd
}

func printHeaders(w io.Writer, header http.Header) error {
	if len(header) == 0 {
		return nil
	}

	keys := make([]string, 0, len(header))
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if _, err := fmt.Fprintf(w, "%s: %s\n", k, strings.Join(header[k], ", ")); err != nil {
			return err
		}
	}
	return nil
}

func buildRequestBody(body, file string, readStdin bool) (io.ReadCloser, error) {
	switch {
	case file != "":
		b, err := os.ReadFile(file)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read body file %s", file)
		}
		return io.NopCloser(bytes.NewReader(b)), nil
	case readStdin:
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, errors.Wrap(err, "failed to read stdin")
		}
		return io.NopCloser(bytes.NewReader(data)), nil
	case body != "":
		return io.NopCloser(strings.NewReader(body)), nil
	default:
		return nil, nil
	}
}

func fetchGatewayRoutes(ctx context.Context, client *http.Client, addr, prefer string) ([]routeOperation, error) {
	listURL := strings.TrimRight(addr, "/") + "/debug/vars/api/list"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build vars list request")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request vars list")
	}
	defer func() { _ = resp.Body.Close() }()

	var varsResp []gatewayVarInfo
	if err := json.NewDecoder(resp.Body).Decode(&varsResp); err != nil {
		return nil, errors.Wrap(err, "failed to decode vars list")
	}

	varName := ""
	if prefer != "" {
		for _, v := range varsResp {
			if v.Name == prefer {
				varName = v.Name
				break
			}
		}
	}
	if varName == "" {
		for _, v := range varsResp {
			if strings.Contains(strings.ToLower(v.Name), "grpc-server-info") {
				varName = v.Name
				break
			}
		}
	}
	if varName == "" && prefer != "" {
		for _, v := range varsResp {
			if strings.Contains(strings.ToLower(v.Name), strings.ToLower(prefer)) {
				varName = v.Name
				break
			}
		}
	}

	if varName == "" {
		names := make([]string, 0, len(varsResp))
		for _, v := range varsResp {
			names = append(names, v.Name)
		}
		return nil, errors.Errorf("gateway vars not found (prefer=%s), available: %v", prefer, names)
	}

	detailURL := strings.TrimRight(addr, "/") + "/debug/vars/api/get/" + url.PathEscape(varName)
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, detailURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build vars get request")
	}

	resp, err = client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to request gateway info")
	}
	defer func() { _ = resp.Body.Close() }()

	var info gatewayInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, errors.Wrap(err, "failed to decode gateway info")
	}

	return info.Method, nil
}

func printRoutes(w io.Writer, routes []routeOperation, prefix string) error {
	if len(routes) == 0 {
		if _, err := fmt.Fprintln(w, "no gateway routes found"); err != nil {
			return err
		}
		return nil
	}

	if _, err := fmt.Fprintf(w, "Registered gateway routes (%d):\n", len(routes)); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "METHOD\tPATH\tOPERATION\tVERB\n"); err != nil {
		return err
	}
	for _, r := range routes {
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", normalizeMethod(r.Method), joinPath(prefix, r.Path), r.Operation, r.Verb); err != nil {
			return err
		}
	}
	return nil
}

var pathPlaceholderRe = regexp.MustCompile(`\{[^}]+\}`)

func applyPathParams(path string, params map[string]string) (string, error) {
	placeholders := pathPlaceholderRe.FindAllString(path, -1)
	for _, ph := range placeholders {
		key := strings.Trim(ph, "{}")
		if idx := strings.IndexByte(key, '='); idx >= 0 {
			key = key[:idx]
		}
		val, ok := params[key]
		if !ok {
			val, ok = params[strings.Trim(ph, "{}")]
		}
		if !ok {
			return "", errors.Errorf("missing path param for %s", ph)
		}
		path = strings.ReplaceAll(path, ph, url.PathEscape(val))
	}
	return path, nil
}

func joinPath(prefix, path string) string {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	prefix = strings.TrimSpace(prefix)
	if prefix == "" || prefix == "/" {
		return path
	}

	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	prefix = strings.TrimSuffix(prefix, "/")
	return prefix + path
}

func normalizeMethod(method string) string {
	if strings.HasPrefix(method, "__") && strings.HasSuffix(method, "__") {
		method = strings.TrimSuffix(strings.TrimPrefix(method, "__"), "__")
	}
	return strings.ToUpper(method)
}

// token helpers
func tokenFilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".lava", "token")
}

func loadToken() (string, error) {
	path := tokenFilePath()
	if path == "" {
		return "", errors.New("cannot resolve home dir for token")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", errors.Wrap(err, "read token")
	}
	return strings.TrimSpace(string(b)), nil
}

func saveToken(tok string) error {
	path := tokenFilePath()
	if path == "" {
		return errors.New("cannot resolve home dir for token")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return errors.Wrap(err, "mkdir token dir")
	}
	return errors.Wrap(os.WriteFile(path, []byte(strings.TrimSpace(tok)), 0o600), "write token")
}

func newLoginCommand() *redant.Command {
	var (
		token   string
		stdin   bool
		fromEnv bool
	)

	return &redant.Command{
		Use:   "login",
		Short: "Save authorization token for subsequent requests",
		Options: redant.OptionSet{
			{Flag: "token", Shorthand: "t", Description: "token string (fallback to stdin)", Value: redant.StringOf(&token)},
			{Flag: "stdin", Description: "read token from stdin", Value: redant.BoolOf(&stdin)},
			{Flag: "env", Description: "read token from LAVA_TOKEN env", Value: redant.BoolOf(&fromEnv)},
		},
		Handler: func(ctx context.Context, inv *redant.Invocation) error {
			defer recovery.Exit()

			if fromEnv && token == "" {
				token = os.Getenv("LAVA_TOKEN")
			}
			if token == "" && stdin {
				b, err := io.ReadAll(inv.Stdin)
				if err != nil {
					return errors.Wrap(err, "read stdin token")
				}
				token = string(b)
			}
			if token == "" && len(inv.Args) > 0 {
				token = inv.Args[0]
			}
			token = strings.TrimSpace(token)
			if token == "" {
				return errors.New("token is required")
			}

			if err := saveToken(token); err != nil {
				return err
			}

			if _, err := fmt.Fprintln(inv.Stdout, "token saved"); err != nil {
				return errors.Wrap(err, "failed to write token saved message")
			}
			return nil
		},
	}
}
