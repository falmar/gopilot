# WIP: gopilot
An attempt to run Chromium automation with bare CDP commands.

> **NOTE:** Breaking changes may occur until the API is finalized.

## Overview
gopilot is my attempt to provide a simple, minimalistic API for automating Chromium browsers. It's not meant to be another Puppeteer. Instead, it's focused on the essential features most users need for straightforward browser tasks—no fluff, just what you need.

## Why Minimalistic?
I wanted to simplify browser automation by sticking to the core functionalities that most of us use:
- Navigation to web pages
- Clicking on elements
- Typing text
- Taking screenshots
- Extracting HTML content

I’ve also added some features for intercepting requests, which is handy if you want to cancel or grab AJAX info. Overall, gopilot aims to be a lightweight tool that doesn’t bog you down with unnecessary complexity.

## Key Features
- **Headfull** mode support: Designed to run as headful and compatible with Docker using Xvfb for display.
- **Headless** mode: Easily switch to headless operation when needed.
- **Navigate** to a specified URL
- **Click** on elements
- **Extract** HTML content from the page
- **Intercept** network requests for those who want to dig deeper
- **Set**, **get**, and **clear** cookies

## Basic Usage Example
Here's a very basic example of how to use gopilot to open a URL:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/falmar/gopilot/pkg/gopilot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := gopilot.NewBrowserConfig()
	b := gopilot.NewBrowser(cfg, logger)

	err := b.Open(ctx, &gopilot.BrowserOpenInput{})
	if err != nil {
		logger.Error("unable to open browser", "error", err)
		return
	}
	defer b.Close(ctx)

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	_, err = page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://www.google.com",
		WaitDomContentLoad: true,
	})
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(2 * time.Second)

	// do some magic ... 
}

```

### More Examples
For more practical illustrations of how to use gopilot, check out the examples provided:

- [Click Element](./examples/click_element/main.go)
- [Cookies](./examples/cookies/main.go)
- [Evaluate JS](./examples/eval/main.go)
- [Listen XHR](./examples/listen_xhr/main.go)
- [Open URL](./examples/open_url/main.go)

### Note on Headless Mode
By default, gopilot runs in headful mode, which may require a display server when running in a Docker container. To switch to headless mode, simply call the `EnableHeadless` method on the `BrowserConfig` object. You can start the browser in headless mode as follows:

```go
// EnableHeadless will make the browser start as headless
cfg := gopilot.NewBrowserConfig()
cfg.EnableHeadless()

// which is basically: 
func (c *BrowserConfig) EnableHeadless() {
	c.AddArgument("--headless=new")
}
```

### TODO:
- Taking screenshots of web pages
- Setting, getting, and clearing local storage
- Typing text into input fields

## Contributions
Contributions are welcome! If you've got a feature request or an idea to share, reach out. Remember to aim for simplicity!
