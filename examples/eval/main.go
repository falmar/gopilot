package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	gopilot2 "github.com/falmar/gopilot/pkg/gopilot"
)

// click cookies with eval

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

	out, err := page.Evaluate(ctx, &gopilot2.PageEvaluateInput{
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
