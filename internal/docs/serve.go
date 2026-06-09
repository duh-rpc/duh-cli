package docs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// DefaultDebounce is the quiet period used to collapse rapid multi-syscall and
// atomic-rename saves into a single reload. It is small so tests run quickly.
const DefaultDebounce = 100 * time.Millisecond

// ServeConfig configures the long-running docs server.
type ServeConfig struct {
	Writer    io.Writer
	SpecPath  string
	Port      int
	PublicURL string
	NoOpen    bool
	// Debounce overrides the reload debounce window; zero uses DefaultDebounce.
	Debounce time.Duration
}

// specState holds the last successfully-parsed spec bytes plus an optional current
// parse error. The served bytes are only ever swapped after a successful parse, so
// a reader never observes half-written content rendered as valid (INV-1). The parse
// error lives in a field distinct from the good bytes, so the two cannot be
// conflated (a parse failure never replaces the good bytes).
type specState struct {
	mu   sync.RWMutex
	good []byte
	err  error
}

func (s *specState) read() (good []byte, parseErr error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.good, s.err
}

// update parses data and swaps the served bytes only on success; on failure it
// retains the last-known-good bytes and records the parse error (INV-1, BC-2).
func (s *specState) update(data []byte) {
	if err := parseSpec(data); err != nil {
		s.mu.Lock()
		s.err = err
		s.mu.Unlock()
		return
	}
	s.mu.Lock()
	s.good = data
	s.err = nil
	s.mu.Unlock()
}

// hub broadcasts reload notifications to every connected SSE client.
type hub struct {
	mu      sync.Mutex
	clients map[chan struct{}]struct{}
}

func newHub() *hub {
	return &hub{clients: make(map[chan struct{}]struct{})}
}

func (h *hub) add() chan struct{} {
	ch := make(chan struct{}, 1)
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *hub) remove(ch chan struct{}) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
}

func (h *hub) broadcast() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}

// Serve starts the docs server for cfg.SpecPath and blocks until ctx is canceled.
// It requires a spec that parses before it binds a port: a missing or unparseable
// spec returns an error and binds nothing (AC #8). On ctx cancellation it shuts the
// HTTP server down cleanly and returns nil (AC #9).
func Serve(ctx context.Context, cfg ServeConfig) error {
	data, err := os.ReadFile(cfg.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to read spec: %w", err)
	}
	if err := parseSpec(data); err != nil {
		return err
	}

	state := &specState{good: data}
	clients := newHub()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, servePage)
	})
	mux.HandleFunc("/redoc.standalone.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		_, _ = w.Write(redocJS)
	})
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		good, parseErr := state.read()
		w.Header().Set("Content-Type", "application/yaml")
		if parseErr != nil {
			w.Header().Set("X-Spec-Parse-Error", parseErr.Error())
		}
		_, _ = w.Write(good)
	})
	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch := clients.add()
		defer clients.remove(ch)
		flusher.Flush()

		for {
			select {
			case <-ctx.Done():
				return
			case <-r.Context().Done():
				return
			case <-ch:
				_, _ = io.WriteString(w, "data: reload\n\n")
				flusher.Flush()
			}
		}
	})

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("failed to bind port: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	url, open := resolveURL(cfg.PublicURL, port, cfg.NoOpen)
	_, _ = fmt.Fprintln(cfg.Writer, url)
	if open {
		_ = openBrowser(url)
	}

	debounce := cfg.Debounce
	if debounce == 0 {
		debounce = DefaultDebounce
	}
	go watch(ctx, cfg.SpecPath, debounce, state, clients)

	srv := &http.Server{Handler: mux}
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// watch watches the directory containing spec and, after a debounce quiet period,
// re-reads and re-parses the file, swaps the served bytes on success, and notifies
// connected browsers. Watching the directory (not the file) survives editors that
// save via atomic rename, which replaces the file's inode.
func watch(ctx context.Context, spec string, debounce time.Duration, state *specState, clients *hub) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer func() { _ = watcher.Close() }()

	dir := filepath.Dir(spec)
	if dir == "" {
		dir = "."
	}
	if err := watcher.Add(dir); err != nil {
		return
	}

	base := filepath.Base(spec)
	timer := time.NewTimer(debounce)
	if !timer.Stop() {
		<-timer.C
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if filepath.Base(event.Name) == base {
				timer.Reset(debounce)
			}
		case <-timer.C:
			data, err := os.ReadFile(spec)
			if err != nil {
				continue
			}
			state.update(data)
			clients.broadcast()
		case _, ok := <-watcher.Errors:
			if !ok {
				return
			}
		}
	}
}
