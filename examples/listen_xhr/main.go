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

	err := b.Open(ctx, &gopilot.BrowserOpenInput{})
	if err != nil {
		logger.Error("unable open page", "error", err)
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

	xMonitor := gopilot.NewXHRMonitor(page)
	ev, err := xMonitor.Listen(ctx, nil)
	if err != nil {
		logger.Error("unable to monitor xhr", "error", err)
		return
	}
	defer xMonitor.Stop(ctx)

	if _, err := page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://www.google.com",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	logger.Info("waiting events")

	timer := time.NewTimer(time.Second * 5)
loop:
	for {
		select {
		case <-timer.C:
			break loop
		case event := <-ev:
			logger.Info("found event", "url", event.URL)
		}
	}
}
