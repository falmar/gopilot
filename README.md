# gopilot

<p align="center">
  <img src="logo/logo.png" alt="GoPilot Logo" width="400"/>
</p>

[![Go Reference](https://pkg.go.dev/badge/github.com/falmar/gopilot.svg)](https://pkg.go.dev/github.com/falmar/gopilot)

A lightweight approach to Chromium automation using basic CDP commands.

> **NOTE:** Breaking changes may occur until the API is finalized.

## Table of Contents
- [Overview](#overview)
- [Why Minimalistic?](#why-minimalistic)
- [Key Features](#key-features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Examples](#examples)
- [Configuration](#configuration)
- [Advanced Usage](#advanced-usage)
- [Project Status & Roadmap](#project-status--roadmap)
- [Contributions](#contributions)

## Overview

gopilot is my attempt to provide a simple, minimalistic API for automating Chromium browsers. It's not meant to be
another Puppeteer. Instead, it's focused on the essential features most users need for straightforward browser tasks—no
fluff, just what you need.

Under the hood gopilot uses [github.com/mafredri/cdp](https://github.com/mafredri/cdp) for chrome communication,
inspired by gRPC provides a really nice and easy API.

## Why Minimalistic?

I wanted to simplify browser automation by sticking to the core functionalities that most of us use:

- Navigation to web pages
- Clicking on elements
- Typing text
- Taking screenshots
- Extracting HTML content

I’ve also added some features for intercepting requests, which is handy if you want to cancel or grab AJAX info.
Overall, gopilot aims to be a lightweight tool that doesn’t bog you down with unnecessary complexity.

## Key Features

- **Headfull** mode support: Designed to run as headful and compatible with Docker using Xvfb for display.
- **Headless** mode: Easily switch to headless operation when needed.
- **Navigate** to a specified URL
- **Query Selector** to find elements on the page
- **Click** on elements
- **Get** and **set** HTML content
- **Intercept Request/Response** (Needs rework in order to allow modifying the request) network requests for those who
  want to dig deeper
- **Set**, **get**, and **clear** cookies and local storage
- **Screenshots** the current page's viewport, the full page or an element's within is bounding box
- **Text Typing** just provide the text to be written, a delay or func can be supplied per keystroke delays 

## Installation

### Prerequisites

- Go 1.24.0 or later
- Chrome or Chromium browser installed on your system

### Installing gopilot

To install gopilot, use the standard Go package installation command:

```bash
go get github.com/falmar/gopilot
```

Import it in your Go code:

```go
import "github.com/falmar/gopilot"
```

## Quick Start

Here's a very basic example of how to use gopilot to open a URL:

```go
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/falmar/gopilot"
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

## Examples

For more practical examples of how to use gopilot, check out the examples provided:

- [Click Element](./examples/click_element/main.go) - Demonstrates how to find and click on elements in a web page
- [Cookies](./examples/cookies/main.go) - Shows how to set, get, and clear cookies
- [Evaluate JS](./examples/eval/main.go) - Examples of executing JavaScript in the browser context
- [Listen XHR](./examples/listen_xhr/main.go) - Demonstrates how to intercept and monitor XHR requests
- [Local Storage](./examples/local_storage/main.go) - Shows how to interact with browser local storage
- [Open Chrome](./examples/open_chrome/main.go) - Basic example of launching a Chrome browser
- [Open URL](./examples/open_url/main.go) - Simple example of navigating to a URL
- [Screenshots](./examples/screenshots/main.go) - Shows how to capture screenshots of pages or elements
- [Search](./examples/search/main.go) - Demonstrates how to search for elements on a page
- [Typing](./examples/typing/main.go) - Examples of typing text into input fields

## Advanced Usage

### Headless Mode

By default, gopilot runs in headful mode, which may require a display server when running in a Docker container. To
switch to headless mode, simply call the `EnableHeadless` method on the `BrowserConfig` object. You can start the
browser in headless mode as follows:

```go
// EnableHeadless will make the browser start as headless
cfg := gopilot.NewBrowserConfig()
cfg.EnableHeadless()
```

## Configuration Options

gopilot provides several configuration options to customize browser behavior:

### Browser Configuration

The `BrowserConfig` struct allows you to configure how the browser is launched:

```go
type BrowserConfig struct {
    // Path specifies the path to the browser executable
    Path string

    // DebugPort specifies the port for debugging connections
    DebugPort string

    // Args contains additional command-line arguments
    Args []string

    // Envs holds environment variables for the browser process
    Envs []string
}
```

### Default Configuration

When you call `gopilot.NewBrowserConfig()`, it creates a configuration with these defaults:

- **Browser Path**: Uses the Chrome executable specified by the `GOPILOT_CHROME_EXECUTABLE` environment variable, or defaults to "google-chrome-stable"
- **Debug Port**: "9222"
- **Default Arguments**: Several arguments for optimal browser operation:
  - `--remote-allow-origins=*`
  - `--no-first-run`
  - `--no-service-autorun`
  - `--no-default-browser-check`
  - `--homepage=about:blank`
  - And several others for stability and performance

### Environment Variables

- **GOPILOT_CHROME_EXECUTABLE**: Set this to specify the path to your Chrome or Chromium executable. For example:
  ```bash
  export GOPILOT_CHROME_EXECUTABLE="/usr/bin/google-chrome"
  ```

### Adding Custom Arguments

You can add custom command-line arguments to the browser:

```go
cfg := gopilot.NewBrowserConfig()
cfg.AddArgument("--disable-gpu")
cfg.AddArgument("--window-size=1280,720")
```


## Project Status & Roadmap

gopilot is currently in active development ("WIP" - Work In Progress). While the core functionality is stable enough for many use cases, the API may change as we refine and improve the library.

### Current Status

- Core browser automation features are implemented and working
- API is functional but may undergo refinements
- Documentation and examples are being expanded

### Planned Features

- Allow users to input an external browser endpoint
- Listen for page/target events to change local data
- Integration tests
- Performance optimizations
- Additional helper methods for common tasks

### Development Priorities

1. API stabilization
2. Improved error handling and recovery
3. Enhanced documentation
4. Performance improvements

## Contributions

Contributions are welcome! If you've got a feature request or an idea to share, reach out.
