package gopilot

type BrowserConfig struct {
	Path      string
	DebugPort string
	Args      []string
}

func NewBrowserConfig() *BrowserConfig {
	c := &BrowserConfig{
		Path:      "/usr/bin/google-chrome-stable",
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

func (c *BrowserConfig) AddArgument(arg string) {
	c.Args = append(c.Args, arg)
}
