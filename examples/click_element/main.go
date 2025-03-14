package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	gopilot2 "github.com/falmar/gopilot/pkg/gopilot"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := gopilot2.NewBrowserConfig()
	b := gopilot2.NewBrowser(cfg, logger)

	if err := b.Open(ctx, &gopilot2.BrowserOpenInput{}); err != nil {
		logger.Error("failed to open browser", "err", err)
		return
	}
	defer b.Close(ctx)

	page, err := b.NewPage(ctx, false)
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	defer page.Close(ctx)

	time.Sleep(time.Second * 2)

	if err := page.Navigate(ctx, "https://www.google.com"); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(time.Second * 2)

	out, err := page.QuerySelector(ctx, &gopilot2.PageQuerySelectorInput{
		Selector: "button#L2AGLb",
	})
	if err != nil {
		logger.Error("unable to query selector", "error", err)
		return
	}

	clickOut, err := out.Element.Click(ctx, &gopilot2.ElementClickInput{
		StepDuration: time.Millisecond * 300,
	})
	if err != nil {
		logger.Error("unable to get rect", "error", err)
		return
	}

	logger.Info("clicked element in position", "x", clickOut.X, "y", clickOut.Y)

	select {
	case <-ctx.Done():
		return
	case <-time.After(time.Second * 60):
	}
}
