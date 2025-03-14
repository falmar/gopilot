package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	gopilot2 "github.com/falmar/gopilot/pkg/gopilot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	cfg := gopilot2.NewBrowserConfig()
	b := gopilot2.NewBrowser(cfg, logger)

	err := b.Open(ctx, &gopilot2.BrowserOpenInput{})
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}

	defer func() {
		if err := b.Close(ctx); err != nil && !strings.Contains(err.Error(), "signal: killed") {
			logger.Error("browser closed", "error", err)
			return
		}
	}()

	p, err := b.NewPage(ctx, false)
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}

	err = p.Navigate(ctx, "https://www.google.com")
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(5 * time.Second)
}
