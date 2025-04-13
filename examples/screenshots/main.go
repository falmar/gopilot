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

	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable to open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	_, err = page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://github.com/falmar/gopilot",
		WaitDomContentLoad: true,
	})
	if err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	time.Sleep(2 * time.Second)

	screenshotOut, err := page.TakeScreenshot(ctx, &gopilot.PageTakeScreenshotInput{
		Format: "png",
		Full:   true,
	})
	if err != nil {
		logger.Error("unable to take page screenshot", "error", err)
		return
	}

	searchOut, err := page.Search(ctx, &gopilot.PageSearchInput{
		Selector: ".Layout-sidebar .BorderGrid-cell",
		Pierce:   true,
	})
	if err != nil {
		logger.Error("unable to find details section", "error", err)
		return
	}

	elScreenshotOut, err := searchOut.Element.TakeScreenshot(ctx, &gopilot.ElementTakeScreenshotInput{
		Format: "png",
	})
	if err != nil {
		logger.Error("unable to take details screenshot", "error", err)
		return
	}

	if err = saveScreenshot("./gopilot_screenshot_fullpage.png", screenshotOut.Data); err != nil {
		logger.Error("unable to save fullpage screenshot")
		return
	}

	if err = saveScreenshot("./gopilot_screenshot_details.png", elScreenshotOut.Data); err != nil {
		logger.Error("unable to save details screenshot")
		return
	}
}

func saveScreenshot(path string, data []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return err
	}

	return f.Sync()
}
