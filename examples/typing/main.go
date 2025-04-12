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
	// Create a context that cancels on interrupt or kill signals.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	// Initialize a logger with debug level logging.
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create browser configuration and initialize a new browser.
	cfg := gopilot.NewBrowserConfig()
	b := gopilot.NewBrowser(cfg, logger)

	// Open the browser.
	if err := b.Open(ctx, &gopilot.BrowserOpenInput{}); err != nil {
		logger.Error("failed to open browser", "err", err)
		return
	}
	defer b.Close(ctx)

	// Sleep for a short duration to wait for the browser to initialize.
	sleep(ctx, time.Second)

	// Create a new page in the browser.
	pOut, err := b.NewPage(ctx, &gopilot.BrowserNewPageInput{})
	if err != nil {
		logger.Error("unable open page", "error", err)
		return
	}
	page := pOut.Page
	defer page.Close(ctx)

	// Sleep briefly to let the page load resources.
	sleep(ctx, time.Second)

	// Navigate to a specified URL and wait for the DOM content to load.
	if _, err := page.Navigate(ctx, &gopilot.PageNavigateInput{
		URL:                "https://wpmtest.org/",
		WaitDomContentLoad: true,
	}); err != nil {
		logger.Error("unable to navigate", "error", err)
		return
	}

	// Interact with the mute button element by searching for its selector.
	srp, err := page.Search(ctx, &gopilot.PageSearchInput{
		Selector: `#speaker_on_off`,
		Pierce:   true,
	})
	if err != nil {
		logger.Error("unable to find cookie accept button")
		return
	}
	_, err = srp.Element.Click(ctx, &gopilot.ElementClickInput{})
	if err != nil {
		logger.Error("unable to remove cookie accept button")
		return
	}

	// Find and interact with a text input element.
	srp, err = page.Search(ctx, &gopilot.PageSearchInput{
		Selector: "#typebox",
	})
	if err != nil {
		logger.Error("unable find input element", "error", err)
		return
	}
	inputEl := srp.Element
	_, err = inputEl.Click(ctx, &gopilot.ElementClickInput{})
	if err != nil {
		logger.Error("unable click input element", "error", err)
		return
	}

	// Sleep before proceeding to interaction with the page.
	sleep(ctx, time.Second)

	// Loop to type words into the input element.
wordLoop:
	for i := 0; i < 20; i++ {
		srp, err := page.Search(ctx, &gopilot.PageSearchInput{
			Selector: "#word-section span.current-word",
		})
		if err != nil {
			logger.Error("unable find current word", "error", err)
			break wordLoop
		}
		currentWord, err := srp.Element.Text(ctx)
		if err != nil {
			logger.Error("unable get current word", "error", err)
			return
		}

		// Add a space after the word as the input resets by pressing "Space".
		_, err = page.TypeText(ctx, &gopilot.PageTypeTextInput{
			Text:  currentWord + " ",
			Delay: time.Millisecond * 250,
		})
		if err != nil {
			logger.Error("unable to type word", "error", err, "word", currentWord)
			return
		}
	}
}

// Sleep function pauses the execution for the specified duration or until the context is cancelled.
func sleep(ctx context.Context, d time.Duration) {
	select {
	case <-ctx.Done():
	case <-time.NewTimer(d).C:
	}
}
