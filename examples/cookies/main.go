package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/falmar/gopilot/pkg/gopilot"
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
		logger.Error("unable open page", "error", err)
		return
	}

	defer func() {
		if err := b.Close(ctx); err != nil && !strings.Contains(err.Error(), "signal: killed") {
			logger.Error("browser closed", "error", err)
			return
		}
	}()

	page, err := b.NewPage(ctx, false)
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}

	if _, err := page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://www.google.com",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(time.Second * 2)

	out, err := page.GetCookies(ctx, &gopilot.GetCookiesInput{})
	if err != nil {
		logger.Error("unable to get cookies", "error", err)
	}

	for _, c := range out.Cookies {
		fmt.Println(c)
	}

	fmt.Println(out.Cookies)
}
