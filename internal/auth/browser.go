package auth

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)


type BrowserOpener struct {
	command  string
	args     []string
	fallback func(url string) error
}


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
	default: 
		
		return &BrowserOpener{
			command:  "xdg-open",
			args:     []string{},
			fallback: tryLinuxFallback,
		}
	}
}


func (bo *BrowserOpener) Open(url string) error {
	
	if os.Getenv("DISPLAY") == "" && runtime.GOOS == "linux" {
		return bo.fallback(url)
	}

	args := append(bo.args, url)
	cmd := exec.Command(bo.command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	
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
		
		return bo.fallback(url)
	}
}


func printManualURL(url string) error {
	fmt.Println("\n┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Please open the following URL in your browser:              │")
	fmt.Println("└─────────────────────────────────────────────────────────────┘")
	fmt.Printf("\n%s\n\n", url)
	fmt.Println("Waiting for authentication...")
	return nil
}


func tryLinuxFallback(url string) error {
	
	cmd := exec.Command("sensible-browser", url)
	if err := cmd.Start(); err == nil {
		return nil
	}

	
	if browser := os.Getenv("BROWSER"); browser != "" {
		cmd := exec.Command(browser, url)
		if err := cmd.Start(); err == nil {
			return nil
		}
	}

	
	return printManualURL(url)
}


func CanOpenBrowser() bool {
	switch runtime.GOOS {
	case "darwin":
		return true
	case "windows":
		return true
	default:
		
		if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
			return true
		}
		return false
	}
}
