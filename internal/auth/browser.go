package auth

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// BrowserOpener handles opening the browser for OAuth flows
type BrowserOpener struct {
	command  string
	args     []string
	fallback func(url string) error
}

// NewBrowserOpener creates a browser opener for the current platform
func NewBrowserOpener() *BrowserOpener {
	switch runtime.GOOS {
	case "darwin":
		return &BrowserOpener{
			command:  "open",
			args:     []string{},
			fallback: printManualURL,
		}
	case "windows":
		return &BrowserOpener{
			command:  "cmd",
			args:     []string{"/c", "start"},
			fallback: printManualURL,
		}
	default: // linux and others
		// Try xdg-open first, then sensible-browser
		return &BrowserOpener{
			command:  "xdg-open",
			args:     []string{},
			fallback: tryLinuxFallback,
		}
	}
}

// Open opens the URL in the default browser
func (bo *BrowserOpener) Open(url string) error {
	// Check if running in a headless environment
	if os.Getenv("DISPLAY") == "" && runtime.GOOS == "linux" {
		return bo.fallback(url)
	}

	args := append(bo.args, url)
	cmd := exec.Command(bo.command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set a timeout for the browser open operation
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return bo.fallback(url)
		}
		return nil
	case <-time.After(5 * time.Second):
		// If browser open takes too long, fall back to manual
		return bo.fallback(url)
	}
}

// printManualURL prints the URL for manual copy-paste
func printManualURL(url string) error {
	fmt.Println("\n┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Please open the following URL in your browser:              │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Printf("\n%s\n\n", url)
	fmt.Println("Waiting for authentication...")
	return nil
}

// tryLinuxFallback tries alternative methods on Linux
func tryLinuxFallback(url string) error {
	// Try sensible-browser
	cmd := exec.Command("sensible-browser", url)
	if err := cmd.Start(); err == nil {
		return nil
	}

	// Try to detect browser from environment
	if browser := os.Getenv("BROWSER"); browser != "" {
		cmd := exec.Command(browser, url)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	// Fall back to printing the URL
	return printManualURL(url)
}

// CanOpenBrowser checks if we can open a browser in the current environment
func CanOpenBrowser() bool {
	switch runtime.GOOS {
	case "darwin":
		return true
	case "windows":
		return true
	default:
		// On Linux, check for DISPLAY or wayland
		if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
			return true
		}
		return false
	}
}
