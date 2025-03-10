package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/falmar/gopilot/internal/gopilot"
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

	p, err := b.NewPage(ctx)
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}

	xMonitor := gopilot.NewXHRMonitor(p)
	ev, err := xMonitor.Listen(ctx, nil)
	if err != nil {
		logger.Error("unable to monitor xhr", "error", err)
		return
	}
	defer xMonitor.Stop(ctx)

	err = p.Navigate(ctx, "https://www.poe2wiki.net/wiki/Path_of_Exile_2_Wiki")
	if err != nil {
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
