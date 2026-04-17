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

	if err := b.Open(ctx, &gopilot.BrowserOpenInput{}); err != nil {
		logger.Error("failed to open browser", "err", err)
		return
	}
	defer b.Close(ctx)

	sleep(ctx, time.Second)

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	sleep(ctx, time.Second)

	if _, err := page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://keyboard-tester.com/",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	sleep(ctx, 2*time.Second)

	// Type "Hello World" using PressKey
	keys := []gopilot.PagePressKeyInput{
		{Key: gopilot.KeyH, Modifiers: gopilot.ModifierShift},
		{Key: gopilot.KeyE},
		{Key: gopilot.KeyL},
		{Key: gopilot.KeyL},
		{Key: gopilot.KeyO},
		{Key: gopilot.KeySpace},
		{Key: gopilot.KeyW, Modifiers: gopilot.ModifierShift},
		{Key: gopilot.KeyO},
		{Key: gopilot.KeyR},
		{Key: gopilot.KeyL},
		{Key: gopilot.KeyD},
		{Key: gopilot.KeyEnter},
	}

	for _, k := range keys {
		if _, err := page.PressKey(ctx, &k); err != nil {
			logger.Error("failed to press key", "error", err, "key", k.Key.Key)
			return
		}
		sleep(ctx, 300*time.Millisecond)
	}

	sleep(ctx, 5*time.Second)
}

func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.NewTimer(d).C:
	}
}
