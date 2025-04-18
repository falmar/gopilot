package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
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

	defer b.Close(ctx)

	time.Sleep(time.Second * 2)

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	time.Sleep(time.Second * 2)

	// SET COOKIES (avoid set cookie pop up)
	_, err = page.SetCookies(ctx, &gopilot.SetCookiesInput{
		Cookies: []gopilot.PageCookie{
			{
				Domain:   ".google.com",
				Name:     "SOCS",
				Value:    "CAISHAgBEhJnd3NfMjAyNTAzMTItMF9SQzIaAmVuIAEaBgiA7-K-Bg",
				Path:     "/",
				Secure:   true,
				HttpOnly: false,
			},
		},
	})
	if err != nil {
		logger.Error("unable to set cookies", "error", err)
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

	// GET COOKIES
	gcOut, err := page.GetCookies(ctx, &gopilot.GetCookiesInput{})
	if err != nil {
		logger.Error("unable to get cookies", "error", err)
	}

	for _, c := range gcOut.Cookies {
		logger.Info("cookie found", "name", c.Name)
	}

	time.Sleep(time.Second * 2)

	// CLEAR COOKIES
	_, err = page.ClearCookies(ctx, &gopilot.ClearCookiesInput{})
	if err != nil {
		logger.Error("unable to clear cookies", "error", err)
		return
	}

	// reload to see accept cookies popup
	_, err = page.Reload(ctx, &gopilot.PageReloadInput{WaitDomContentLoad: true})
	if err != nil {
		logger.Error("unable to reload page", "error", err)
		return
	}
	time.Sleep(time.Second * 2)
}
