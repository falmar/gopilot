package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/falmar/gopilot/pkg/gopilot"
)

// Keep chrome open with default setting for local raw cdp testing
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
		logger.Error("unable open page", "error", err)
		return
	}
	defer b.Close(ctx)

	select {
	case <-ctx.Done():
		return
	}
}
