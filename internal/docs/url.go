package docs

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// resolveURL determines the URL to print for the running server and whether to
// attempt opening a browser, following the order documented in the blueprint:
//
//  1. --public-url set       → print it; do not open a browser.
//  2. WEB_HOST set (Google   → print https://<port>-<WEB_HOST>; do not open.
//     Cloud Workstation)
//  3. otherwise (local)      → print http://localhost:<port>; open unless --no-open.
func resolveURL(publicURL string, port int, noOpen bool) (url string, open bool) {
	if publicURL != "" {
		return publicURL, false
	}
	if host := os.Getenv("WEB_HOST"); host != "" {
		return fmt.Sprintf("https://%d-%s", port, host), false
	}
	return fmt.Sprintf("http://localhost:%d", port), !noOpen
}

// openBrowser launches the OS browser opener for url. Failures are returned but
// are non-fatal to serving; the URL is always printed regardless.
func openBrowser(url string) error {
	var name string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		name = "open"
		args = []string{url}
	case "windows":
		name = "cmd"
		args = []string{"/c", "start", url}
	default:
		name = "xdg-open"
		args = []string{url}
	}
	return exec.Command(name, args...).Start()
}
