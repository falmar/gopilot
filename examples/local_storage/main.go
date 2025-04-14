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

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{
		NewTab: true,
	})
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	_, err = page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL: "https://example.com",
	})
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(time.Second)

	_, err = page.SetLocalStorage(ctx, &gopilot.SetLocalStorageInput{
		Items: []gopilot.LocalStorageItem{
			gopilot.LocalStorageItem{
				Name:  "gopilot_exampleItem",
				Value: "hello world",
			},
		},
	})
	if err != nil {
		logger.Error("unable to set local storage", "error", err)
		return
	}

	lsOut, err := page.GetLocalStorage(ctx, &gopilot.GetLocalStorageInput{})
	if err != nil {
		logger.Error("unable to get local storage", "error", err)
		return
	}

	logger.Info("local storage found", "items", lsOut.Items)

	err = page.ClearLocalStorage(ctx)
	if err != nil {
		logger.Error("unable to clear local storage", "error", err)
		return
	}

	logger.Info("local storage cleared!")
}
