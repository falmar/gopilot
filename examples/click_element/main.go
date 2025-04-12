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
		Level: slog.LevelInfo,
	}))

	cfg := gopilot.NewBrowserConfig()
	b := gopilot.NewBrowser(cfg, logger)

	if err := b.Open(ctx, &gopilot.BrowserOpenInput{}); err != nil {
		logger.Error("failed to open browser", "err", err)
		return
	}
	defer b.Close(ctx)

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	time.Sleep(time.Second * 2)

	if _, err := page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://cps-check.com/mouse-buttons-test",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(time.Second * 2)

	// or use page.QuerySelector
	out, err := page.Search(ctx, &gopilot.PageSearchInput{
		Selector: "#mouse-container",
		Pierce:   true,
	})
	if err != nil {
		logger.Error("unable to query selector", "error", err)
		return
	}

	// EXAMPLE SIMPLE CLICK
	clickOut, err := out.Element.Click(ctx, &gopilot.ElementClickInput{
		StepDuration: time.Millisecond * 300,
	})
	if err != nil {
		logger.Error("unable to click", "error", err)
		return
	}
	logger.Info("clicked element in position", "x", clickOut.X, "y", clickOut.Y)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second * 3):
	}

	// EXAMPLE HOLD CLICK
	holdDuration := time.Second * 5
	logger.Info("clicking and holding", "duration", holdDuration)
	clickOut, err = out.Element.Click(ctx, &gopilot.ElementClickInput{
		HoldDuration: holdDuration,
	})
	if err != nil {
		logger.Error("unable to click", "error", err)
		return
	}
	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second * 2):
	}

	// EXAMPLE HOLD CLICK WITH USER RELEASE
	releaseDuration := time.Second * 5
	logger.Info("clicking and holding with release handle", "duration", releaseDuration)
	clickOut, err = out.Element.Click(ctx, &gopilot.ElementClickInput{
		ReturnHoldRelease: true,
	})
	if err != nil {
		logger.Error("unable to click", "error", err)
		return
	}

	time.Sleep(releaseDuration)
	if err = clickOut.Release(); err != nil {
		logger.Error("release handle error", "error", err)
		return
	}

	logger.Info("done")
}
