package servecmd

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/pubgo/funk/recovery"
	cli "github.com/urfave/cli/v3"
)

// customFileServer wraps the standard FileServer to add custom headers and ETag support
type customFileServer struct {
	root http.FileSystem
	fs   http.Handler
}

func (s *customFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := path.Clean(r.URL.Path)

	// Handle special health check endpoints
	if urlPath == "/healthz" || urlPath == "/ping" {
		if urlPath == "/healthz" {
			w.Write([]byte("OK"))
		} else {
			w.Write([]byte("pong"))
		}
		return
	}

	// Check if the path exists and get file info
	f, err := s.root.Open(urlPath)
	if err != nil {
		s.fs.ServeHTTP(w, r)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		s.fs.ServeHTTP(w, r)
		return
	}

	// Generate ETag based on file size and modification time
	etag := generateETag(fi)
	w.Header().Set("ETag", etag)

	// Check if client sent If-None-Match header
	if match := r.Header.Get("If-None-Match"); match != "" {
		if match == etag {
			// Resource not modified, return 304
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}
	// Let the standard FileServer handle the rest
	s.fs.ServeHTTP(w, r)
}

// generateETag creates an ETag value based on file modification time and size
func generateETag(fi os.FileInfo) string {
	modTime := fi.ModTime().UTC().Format(time.RFC3339Nano)
	size := fi.Size()
	hash := sha256.New()
	fmt.Fprintf(hash, "%s-%d", modTime, size)
	return fmt.Sprintf(`"%x"`, hash.Sum(nil))
}

func New() *cli.Command {
	var (
		flagVerbose bool
	)

	return &cli.Command{
		Name:  "serve",
		Usage: "serve `pwd` via http at *:8080",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "verbose",
				Usage:       "verbose",
				Local:       true,
				Value:       flagVerbose,
				Destination: &flagVerbose,
			},
		},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			wd, err := os.Getwd()
			if err != nil {
				panic(err)
			}

			port := "8080"
			if p := os.Getenv("PORT"); p != "" {
				port = p
			}
			fmt.Printf("http://localhost:%v/\n", port)

			fileSystem := http.Dir(wd)
			fileServer := http.FileServer(fileSystem)
			var handler http.Handler = &customFileServer{root: fileSystem, fs: fileServer}

			if flagVerbose {
				h := handler
				handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					log.Printf("%s %s", r.Method, r.URL.Path)
					h.ServeHTTP(w, r)
				})
			}
			panic(http.ListenAndServe(fmt.Sprintf(":%v", port), handler))
		},
	}
}
