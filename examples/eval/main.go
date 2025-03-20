package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	gopilot "github.com/falmar/gopilot/pkg/gopilot"
)

// click cookies with eval

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

	out, err := page.Evaluate(ctx, &gopilot.PageEvaluateInput{
		ReturnValue:  true,
		AwaitPromise: false,
		Expression: `(() => {
const button = document.querySelector('button#L2AGLb')
button.click()

return 'button was clicked and the className is: ' + button.className.toString()

})()`,
	})
	if err != nil {
		logger.Error("unable to evaluate page", "error", err)
		return
	}

	logger.Info("evaluated page", "value", string(out.Value))
}
