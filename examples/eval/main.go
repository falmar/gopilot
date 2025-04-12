package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/falmar/gopilot"
)

// click cookies with eval

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

	time.Sleep(time.Second * 2)

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

	time.Sleep(time.Second * 2)

	logger.Info("evaluated page", "value", string(out.Value))

	out, err = page.Evaluate(ctx, &gopilot.PageEvaluateInput{
		ReturnValue:  true,
		AwaitPromise: true,
		Expression: `
(async function() {
    const textArea = document.querySelector("textarea#APjFqb");
    const message = "Do you want to eat pizza?";

    // Clear the textarea
    textArea.value = '';

    // Function to simulate typing
    for (let char of message) {
        textArea.value += char; // Add one character at a time
        textArea.dispatchEvent(new Event('input', { bubbles: true })); // Trigger input event
        await new Promise(resolve => setTimeout(resolve, Math.random() * (200 - 70) + 70)); // Wait 70-200 ms
    }

    // Wait for 1 seconds before pressing enter
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    // Trigger the enter key
    const event = new KeyboardEvent('keydown', {
        key: 'Enter',
        keyCode: 13,
        code: 'Enter',
        which: 13,
        bubbles: true
    });
    textArea.dispatchEvent(event);
})();
`,
	})
	if err != nil {
		logger.Error("unable to evaluate page", "error", err)
		return
	}

	time.Sleep(time.Second * 2)
}
