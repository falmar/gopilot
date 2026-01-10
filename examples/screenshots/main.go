package main

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/falmar/gopilot"
)

var pageExampleHTML string = `
<!doctype html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>IMAGE_TYPE Screenshot</title>
    <style>
        img {
            max-width: 90%;
            height: auto;
        }
    </style>
</head>
<body>
<h1>IMAGE_TYPE screenshot:</h1>
<img src="data:image/png;base64,BASE64_IMAGE_DATA" alt="Screenshot">
</body>
</html>
`

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

	time.Sleep(time.Second * 2)

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

	time.Sleep(time.Second * 2)

	pageDisplayScreenshot, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{
		NewTab: true,
	})
	if err != nil {
		logger.Error("unable to open page for example", "error", err)
		return
	}

	pageContent := strings.Replace(pageExampleHTML, "IMAGE_TYPE", "Full page", -1)
	pageContent = strings.Replace(pageContent,
		"BASE64_IMAGE_DATA",
		base64.RawStdEncoding.EncodeToString(screenshotOut.Data),
		-1)
	err = pageDisplayScreenshot.Page.SetContent(ctx, pageContent)
	if err != nil {
		logger.Error("unable to open page for example", "error", err)
		return
	}

	time.Sleep(time.Second * 2)

	pageContent = strings.Replace(pageExampleHTML, "IMAGE_TYPE", "Details section", -1)
	pageContent = strings.Replace(pageContent,
		"BASE64_IMAGE_DATA",
		base64.RawStdEncoding.EncodeToString(elScreenshotOut.Data),
		-1)
	err = pageDisplayScreenshot.Page.SetContent(ctx, pageContent)
	if err != nil {
		logger.Error("unable to open page for example", "error", err)
		return
	}

	time.Sleep(time.Second * 5)
}
