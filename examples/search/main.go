package main

import (
	"context"
	"errors"
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

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{NewTab: true})
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	_, err = page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://falmar.github.io/gopilot/examples/search/tpl.html",
		WaitDomContentLoad: true,
	})
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	srp, err := page.Search(ctx, &gopilot.PageSearchInput{
		Selector:     "body button.green",
		Pierce:       true,
		WaitDuration: time.Second * 5,
		TickDuration: time.Millisecond * 250,
	})
	if errors.Is(err, gopilot.ErrElementSearchTimeout) || errors.Is(err, gopilot.ErrElementNotFound) {
		logger.Error("element not found", "error", err)
		return
	} else if err != nil {
		logger.Error("page search error")
		return
	}

	_, err = srp.Element.Click(ctx, &gopilot.ElementClickInput{})
	if err != nil {
		logger.Error("unable to click", "error", err)
		return
	}

	time.Sleep(time.Second)
}
