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

	// Connect to an existing Chrome instance instead of launching a new one
	// To start Chrome with remote debugging:
	//   google-chrome --remote-debugging-port=9222
	// Then use the WebSocket URL from Chrome's output, or the simplified HTTP format below

	cfg := gopilot.NewBrowserConfig()
	cfg.ConnectionURL = "http://localhost:9222" // Connect to existing browser

	b := gopilot.NewBrowser(cfg, logger)

	err := b.Open(ctx, &gopilot.BrowserOpenInput{})
	if err != nil {
		logger.Error("unable to connect to browser", "error", err)
		logger.Info("make sure Chrome is running with: google-chrome --remote-debugging-port=9222")
		return
	}
	// Note: Close() will close pages but NOT kill the external browser
	defer b.Close(ctx)

	logger.Info("successfully connected to external browser")

	// Show all pages in browser (including existing tabs)
	allPagesOut, err := b.GetAllPages(ctx, &gopilot.BrowserGetPagesInput{})
	if err != nil {
		logger.Error("unable to get all pages", "error", err)
		return
	}
	logger.Info("found existing pages in browser", "count", len(allPagesOut.Pages))

	// Create a new session page
	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{NewTab: true})
	if err != nil {
		logger.Error("unable to create page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	// Navigate to a URL
	_, err = page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://super.walmart.com.mx",
		WaitDomContentLoad: true,
	})
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	logger.Info("successfully navigated to example.com")

	// Show difference between session pages and all pages
	sessionPagesOut, err := b.GetPages(ctx, &gopilot.BrowserGetPagesInput{})
	if err == nil {
		logger.Info("session pages (will be closed)", "count", len(sessionPagesOut.Pages))
	}

	time.Sleep(3 * time.Second)

	logger.Info("closing connection (only session pages closed, browser remains open)")
}
