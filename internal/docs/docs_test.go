package docs_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	duh "github.com/duh-rpc/duh-cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validSpec is a minimal OpenAPI 3.0 document with a distinctive title and path
// used to grep for the inlined spec in served and exported output.
const validSpec = `openapi: 3.0.0
info:
  title: Widget API
  version: 1.0.0
paths:
  /widgets.create:
    post:
      operationId: createWidget
      responses:
        '200':
          description: OK
`

// validSpecV2 is a second valid spec used to verify reload/recovery picks up
// changed content.
const validSpecV2 = `openapi: 3.0.0
info:
  title: Gadget API
  version: 2.0.0
paths:
  /gadgets.create:
    post:
      operationId: createGadget
      responses:
        '200':
          description: OK
`

const invalidSpec = `:::: this is not valid yaml openapi {{{`

// syncBuffer is a goroutine-safe io.Writer so a test can poll stdout while serve
// writes to it from another goroutine.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

// startServe runs `docs serve --port 0` in a goroutine and returns the resolved
// base URL printed on stdout, plus a channel that yields the process exit code.
func startServe(t *testing.T, ctx context.Context, specPath string, extraArgs ...string) (string, *syncBuffer, chan int) {
	t.Helper()
	stdout := &syncBuffer{}
	exit := make(chan int, 1)

	args := append([]string{"docs", "serve", specPath, "--port", "0", "--no-open"}, extraArgs...)
	go func() {
		exit <- duh.RunCmd(ctx, stdout, args)
	}()

	var url string
	require.Eventually(t, func() bool {
		line := strings.TrimSpace(strings.SplitN(stdout.String(), "\n", 2)[0])
		if strings.HasPrefix(line, "http://") || strings.HasPrefix(line, "https://") {
			url = line
			return true
		}
		return false
	}, 5*time.Second, 10*time.Millisecond)

	return url, stdout, exit
}

// --- export ---

// AC #1, #2, BC-1: export writes a single self-contained HTML file with the spec
// and ReDoc inlined and no external asset references.
func TestDocsExportSelfContained(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	output := filepath.Join(dir, "docs.html")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "export", specPath, "-o", output})
	require.Equal(t, 0, exitCode)

	data, err := os.ReadFile(output)
	require.NoError(t, err)
	html := string(data)

	// Inlined spec content present (AC #2).
	assert.Contains(t, html, "Widget API")
	assert.Contains(t, html, "/widgets.create")
	// Inlined ReDoc bundle + bootstrap present (AC #1, #2).
	assert.Contains(t, html, "Redoc.init")
	assert.Contains(t, html, "For license information please see redoc.standalone.js")
	// No external asset references (BC-1).
	assert.NotContains(t, html, `src="http`)
	assert.NotContains(t, html, `href="http`)
}

// AC #1: default output path is docs.html.
func TestDocsExportDefaultOutput(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	defer func() { _ = os.Chdir(wd) }()

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "export", "spec.yaml"})
	require.Equal(t, 0, exitCode)

	_, err = os.Stat(filepath.Join(dir, "docs.html"))
	require.NoError(t, err)
}

// AC #3: export on an unparseable spec exits 2, prints an error, and does not
// create the output file.
func TestDocsExportInvalidSpecNoOutput(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.yaml")
	output := filepath.Join(dir, "docs.html")
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "export", specPath, "-o", output})

	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error")

	_, err := os.Stat(output)
	assert.True(t, os.IsNotExist(err))
}

// AC #3 (INV-2/BC-3): export on an unparseable spec does not modify a pre-existing
// good output file.
func TestDocsExportInvalidSpecDoesNotClobber(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.yaml")
	output := filepath.Join(dir, "docs.html")
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))
	require.NoError(t, os.WriteFile(output, []byte("PREVIOUS GOOD CONTENT"), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "export", specPath, "-o", output})
	assert.Equal(t, 2, exitCode)

	data, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.Equal(t, "PREVIOUS GOOD CONTENT", string(data))
}

// AC #4: export over an existing docs.html replaces it silently and leaves no
// temp file behind.
func TestDocsExportOverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	output := filepath.Join(dir, "docs.html")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))
	require.NoError(t, os.WriteFile(output, []byte("OLD CONTENT"), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "export", specPath, "-o", output})
	require.Equal(t, 0, exitCode)

	data, err := os.ReadFile(output)
	require.NoError(t, err)
	assert.NotEqual(t, "OLD CONTENT", string(data))
	assert.Contains(t, string(data), "Widget API")

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	for _, e := range entries {
		assert.False(t, strings.HasSuffix(e.Name(), ".tmp"))
	}
}

// --- serve ---

// AC #5: serve --port 0 binds a free port, prints the resolved URL, and exposes
// /, /openapi.yaml, and /events.
func TestDocsServeEndpoints(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url, _, exit := startServe(t, ctx, specPath)

	require.True(t, strings.HasPrefix(url, "http://localhost:"))

	// GET / returns HTML referencing the embedded ReDoc with no external CDN refs (BC-1).
	index := httpGet(t, url+"/")
	assert.Contains(t, index, "redoc.standalone.js")
	assert.NotContains(t, index, `src="http`)
	assert.NotContains(t, index, `href="http`)

	// GET /openapi.yaml returns the current spec bytes.
	spec := httpGet(t, url+"/openapi.yaml")
	assert.Contains(t, spec, "Widget API")

	// GET /events is the SSE endpoint.
	resp, err := http.Get(url + "/events")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

	cancel()
	assert.Equal(t, 0, <-exit)
}

// AC #6: after the spec file changes on disk, an SSE client receives a reload
// event within the debounce window.
func TestDocsServeReloadEvent(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url, _, _ := startServe(t, ctx, specPath)

	events := subscribeSSE(t, ctx, url+"/events")

	// Modify the spec on disk; expect a reload notification.
	require.NoError(t, os.WriteFile(specPath, []byte(validSpecV2), 0644))

	require.Eventually(t, func() bool {
		return strings.Contains(events.String(), "reload")
	}, 5*time.Second, 20*time.Millisecond)

	// The served spec reflects the new content (in-place re-render source).
	require.Eventually(t, func() bool {
		return strings.Contains(httpGet(t, url+"/openapi.yaml"), "Gadget API")
	}, 5*time.Second, 20*time.Millisecond)
}

// AC #7 (INV-1, BC-2): replacing the spec with invalid content does not crash the
// server; the spec endpoint surfaces the parse error and keeps serving the
// last-known-good bytes; a later valid save recovers.
func TestDocsServeInvalidSaveRecovers(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	url, _, exit := startServe(t, ctx, specPath)

	// Write invalid content.
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

	// The endpoint surfaces the parse error while keeping the last-known-good bytes.
	require.Eventually(t, func() bool {
		resp, err := http.Get(url + "/openapi.yaml")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.Header.Get("X-Spec-Parse-Error") != "" && strings.Contains(string(body), "Widget API")
	}, 5*time.Second, 20*time.Millisecond)

	// A later valid save recovers: the error clears and new content is served.
	require.NoError(t, os.WriteFile(specPath, []byte(validSpecV2), 0644))
	require.Eventually(t, func() bool {
		resp, err := http.Get(url + "/openapi.yaml")
		if err != nil {
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		return resp.Header.Get("X-Spec-Parse-Error") == "" && strings.Contains(string(body), "Gadget API")
	}, 5*time.Second, 20*time.Millisecond)

	// Server is still alive and shuts down cleanly.
	cancel()
	assert.Equal(t, 0, <-exit)
}

// AC #8: serve on a missing spec exits 2 and binds no port (returns immediately).
func TestDocsServeMissingSpecExits2(t *testing.T) {
	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "serve", "does-not-exist.yaml", "--port", "0", "--no-open"})
	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error")
}

// AC #8: serve on an unparseable spec at startup exits 2 and binds no port.
func TestDocsServeInvalidSpecExits2(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "bad.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(invalidSpec), 0644))

	var stdout bytes.Buffer
	exitCode := duh.RunCmd(context.Background(), &stdout, []string{"docs", "serve", specPath, "--port", "0", "--no-open"})
	assert.Equal(t, 2, exitCode)
	assert.Contains(t, stdout.String(), "Error")
}

// AC #9: serve shuts down cleanly on context cancellation and exits 0.
func TestDocsServeCleanShutdown(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	ctx, cancel := context.WithCancel(context.Background())
	_, _, exit := startServe(t, ctx, specPath)

	cancel()
	select {
	case code := <-exit:
		assert.Equal(t, 0, code)
	case <-time.After(5 * time.Second):
		t.Fatal("serve did not shut down within 5s")
	}
}

// AC #10: with WEB_HOST set, serve prints https://<port>-<WEB_HOST> and does not
// open a browser (the WEB_HOST branch never opens).
func TestDocsServeWebHostURL(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	t.Setenv("WEB_HOST", "ws.cluster.cloudworkstations.dev")

	stdout := &syncBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exit := make(chan int, 1)
	go func() {
		exit <- duh.RunCmd(ctx, stdout, []string{"docs", "serve", specPath, "--port", "0"})
	}()

	var url string
	require.Eventually(t, func() bool {
		url = strings.TrimSpace(strings.SplitN(stdout.String(), "\n", 2)[0])
		return strings.HasPrefix(url, "https://")
	}, 5*time.Second, 10*time.Millisecond)

	assert.Regexp(t, `^https://\d+-ws\.cluster\.cloudworkstations\.dev$`, url)

	cancel()
	assert.Equal(t, 0, <-exit)
}

// AC: --public-url overrides the printed URL.
func TestDocsServePublicURL(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "openapi.yaml")
	require.NoError(t, os.WriteFile(specPath, []byte(validSpec), 0644))

	stdout := &syncBuffer{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	exit := make(chan int, 1)
	go func() {
		exit <- duh.RunCmd(ctx, stdout, []string{"docs", "serve", specPath, "--port", "0", "--public-url", "https://docs.example.com"})
	}()

	require.Eventually(t, func() bool {
		return strings.Contains(stdout.String(), "https://docs.example.com")
	}, 5*time.Second, 10*time.Millisecond)

	cancel()
	assert.Equal(t, 0, <-exit)
}

// --- helpers ---

func httpGet(t *testing.T, url string) string {
	t.Helper()
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return string(body)
}

// subscribeSSE connects to an SSE endpoint and accumulates streamed bytes into a
// goroutine-safe buffer the caller can poll.
func subscribeSSE(t *testing.T, ctx context.Context, url string) *syncBuffer {
	t.Helper()
	buf := &syncBuffer{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	go func() {
		defer func() { _ = resp.Body.Close() }()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			_, _ = fmt.Fprintln(buf, scanner.Text())
		}
	}()

	return buf
}
