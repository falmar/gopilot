package gopilot

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
}

// NewBrowserConfig creates a new BrowserConfig with default settings.
// The default Path is "google-chrome-stable" and the default DebugPort is "9222".
// It includes several default command-line arguments for browser startup.
func NewBrowserConfig() *BrowserConfig {
	c := &BrowserConfig{
		Path:      "google-chrome-stable", // can be changed by user
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
