package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	gopilot "github.com/falmar/gopilot/pkg/gopilot"
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
		URL:                "https://www.google.com",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	out, err := page.QuerySelector(ctx, &gopilot.PageQuerySelectorInput{
		Selector: "button#L2AGLb",
	})
	if err != nil {
		logger.Error("unable to query selector", "error", err)
		return
	}

	textContent, err := out.Element.Text(ctx)
	if err != nil {
		logger.Error("unable to get text", "error", err)
		return
	}
	logger.Info("button text", "text", textContent)
	time.Sleep(time.Second * 1)

	clickOut, err := out.Element.Click(ctx, &gopilot.ElementClickInput{
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
	case <-time.After(time.Second * 5):
	}
}
