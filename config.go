package gopilot

import (
	"os"
	"time"
)

var defaultOpenTimeout = 5 * time.Second
var defaultCloseTimeout = 5 * time.Second
var defaultClosePause = 50 * time.Millisecond

// BrowserConfig holds configuration settings for launching a browser instance.
type BrowserConfig struct {
	// Path specifies the path to the browser executable.
	Path string

	// DebugPort specifies the port for debugging connections.
	DebugPort string

	// Args contains additional command-line arguments to pass when launching the browser.
	Args []string

	// Envs holds any environment variables to set for the browser process.
	Envs []string

	// OpenTimeout defines how long to wait for Chrome to print the "DevTools listening on" message during startup.
	// If nil, a default of 5 seconds is used. Increase this if your environment starts Chrome slowly.
	OpenTimeout *time.Duration

	// CloseTimeout defines how long to wait for the Chrome process to terminate during shutdown.
	// If nil, a default of 5 seconds is used. Increase this if your environment needs more time to exit cleanly.
	CloseTimeout *time.Duration

	// ConnectionURL specifies the URL of an existing Chrome/Chromium browser to connect to.
	// When set, gopilot will connect to the existing browser instead of launching a new process.
	// Supports both WebSocket URLs (ws://127.0.0.1:9222/devtools/browser/UUID) and HTTP (http://127.0.0.1:9222).
	// The external browser will NOT be closed when Browser.Close() is called.
	//
	// Example:
	//   cfg := gopilot.NewBrowserConfig()
	//   cfg.ConnectionURL = "http://127.0.0.1:9222"
	ConnectionURL string
}

// NewBrowserConfig creates a new BrowserConfig with default settings.
// The default Path is "chromium" and the default DebugPort is "9222".
// It includes several default command-line arguments for browser startup.
func NewBrowserConfig() *BrowserConfig {
	execPath := os.Getenv("GOPILOT_CHROME_EXECUTABLE")
	if execPath == "" {
		execPath = "chromium"
	}

	c := &BrowserConfig{
		Path:      execPath, // can be changed by user
		DebugPort: "9222",
		Args: []string{
			"--remote-allow-origins=*",
			"--no-first-run",
			"--no-service-autorun",
			"--no-default-browser-check",
			"--homepage=about:blank",
			"--no-pings",
			"--password-store=basic",
			"--disable-infobars",
			"--disable-breakpad",
			"--disable-dev-shm-usage",
			"--disable-session-crashed-bubble",
			"--disable-search-engine-choice-screen",
			"--window-size=1920,1080",
		},
	}
	return c
}

// AddArgument appends an additional command-line argument to the browser configuration.
// This allows users to customize the launch options for the browser instance.
func (c *BrowserConfig) AddArgument(arg string) {
	c.Args = append(c.Args, arg)
}

// EnableHeadless will make the browser to start as headless
func (c *BrowserConfig) EnableHeadless() {
	c.AddArgument("--headless=new")
}
