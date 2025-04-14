package gopilot

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/mafredri/cdp/devtool"
)

// Browser defines a contract for browser operations.
// It allows managing browser instances and interacting with web pages.
type Browser interface {
	// Open initiates a new browser session.
	// It takes a context and BrowserOpenInput as parameters.
	// Returns an error if the browser fails to start.
	Open(ctx context.Context, in *BrowserOpenInput) error

	// NewPage creates a new page or tab in the browser.
	// Accepts context and BrowserNewPageInput to specify creation parameters.
	// Returns a BrowserNewPageOutput containing the newly created page
	// or an error if the page cannot be created.
	NewPage(ctx context.Context, in *BrowserNewPageInput) (*BrowserNewPageOutput, error)

	// GetPages retrieves all active pages in the current browser session.
	// Requires a context and BrowserGetPagesInput for the request.
	// Returns a BrowserGetPagesOutput with a list of pages or an error
	// if retrieving the pages fails.
	GetPages(ctx context.Context, in *BrowserGetPagesInput) (*BrowserGetPagesOutput, error)

	// Close shuts down the browser instance and cleans up any resources.
	// It takes a context and returns an error if the browser fails to close.
	Close(ctx context.Context) error

	// GetDevToolClient retrieves the DevTools client associated with the browser.
	// This client allows for advanced interactions with the browser's DevTools protocol,
	// enabling custom actions and low-level debugging or profiling features.
	GetDevToolClient() *devtool.DevTools
}

type browser struct {
	config   *BrowserConfig
	logger   *slog.Logger
	instance *exec.Cmd
	datadir  string
	mux      sync.RWMutex
	devtool  *devtool.DevTools
	pages    []Page
	waitChan chan error
}

// NewBrowser creates a new browser instance with the given configuration and logger.
func NewBrowser(cfg *BrowserConfig, logger *slog.Logger) Browser {
	return &browser{
		config:   cfg,
		logger:   logger,
		pages:    make([]Page, 0),
		waitChan: make(chan error),
	}
}

// BrowserOpenInput contains parameters required to open a browser.
type BrowserOpenInput struct{}

// Open initializes and starts the browser process.
func (b *browser) Open(ctx context.Context, in *BrowserOpenInput) error {
	tempDir, err := os.MkdirTemp("", "gopilot")
	if err != nil {
		return err
	}
	b.datadir = tempDir
	b.logger.Debug("created data dir", "path", b.datadir)

	b.instance = exec.Command(b.config.Path)
	b.instance.Env = b.config.Envs
	b.instance.Args = append(
		b.config.Args,
		fmt.Sprintf("--user-data-dir=%s", tempDir),
	)

	// TODO: check if debug port is already in use (when unset)
	// in order to use next one incrementally
	if b.config.DebugPort != "" {
		b.instance.Args = append(
			b.instance.Args,
			fmt.Sprintf("--remote-debugging-port=%s", b.config.DebugPort),
		)
	}

	// Handle stderr to capture DevTools URL
	dtChan := make(chan string)
	stderr, err := b.instance.StderrPipe()
	if err != nil {
		return err
	}
	defer stderr.Close()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "DevTools listening on") {
				dtChan <- line
			}
			b.logger.Debug("chromesdterr", "msg", line)
		}
	}()

	err = b.instance.Start()
	if err != nil {
		return err
	}
	b.logger.Debug("waiting for devtool url message")

	go func() {
		b.waitChan <- b.instance.Wait()
	}()

	// Wait for the DevTools URL message or timeout
	waitDuration := time.Second * 5
	var devtoolsURLString string
	select {
	case err := <-b.waitChan:
		return fmt.Errorf("exec wait exited unexpectedly or too soon: %w", err)
	case <-time.NewTimer(waitDuration).C:
		return fmt.Errorf("duration %s exceeded waiting for devtool url", waitDuration)

	// successful case
	case dtMessage := <-dtChan:
		dtSplit := strings.Split(dtMessage, "DevTools listening on")
		if len(dtSplit) < 2 {
			return errors.New("unable to obtain dev tool url")
		}
		devtoolsURLString = strings.TrimSpace(dtSplit[1])
	}
	devtoolURL, err := url.Parse(devtoolsURLString)
	if err != nil {
		return err
	}

	b.logger.Debug("creating devtool", "url", devtoolsURLString)
	b.devtool = devtool.New(fmt.Sprintf("http://127.0.0.1:%s", devtoolURL.Port()))

	return nil
}

// BrowserNewPageInput contains parameters for creating a new page.
type BrowserNewPageInput struct {
	NewTab bool
	URL    string
}

// BrowserNewPageOutput contains the result of creating a new page.
type BrowserNewPageOutput struct {
	Page Page
}

// NewPage creates a new tab or page in the browser.
func (b *browser) NewPage(ctx context.Context, in *BrowserNewPageInput) (*BrowserNewPageOutput, error) {
	b.logger.Debug("creating new page cdp target")

	var t *devtool.Target
	var err error

	if in.NewTab {
		if in.URL == "" {
			t, err = b.devtool.Create(ctx)
		} else {
			t, err = b.devtool.CreateURL(ctx, in.URL)
		}
	} else {
		t, err = b.devtool.Get(ctx, devtool.Page)
	}

	if err != nil {
		return nil, err
	}

	p, err := newPage(ctx, t, b.logger)
	if err != nil {
		return nil, err
	}

	b.mux.Lock()
	b.pages = append(b.pages, p)
	b.mux.Unlock()

	return &BrowserNewPageOutput{Page: p}, nil
}

// BrowserGetPagesInput represents parameters to obtain open pages.
type BrowserGetPagesInput struct{}

// BrowserGetPagesOutput contains the list of open browser pages.
type BrowserGetPagesOutput struct {
	Pages []Page
}

// GetPages retrieves the list of active pages in the browser.
func (b *browser) GetPages(ctx context.Context, _ *BrowserGetPagesInput) (*BrowserGetPagesOutput, error) {
	var pg []Page

	b.mux.Lock()
	defer b.mux.Unlock()

	// Filter out closed pages
	for _, p := range b.pages {
		if p.(*page).closed {
			continue
		}
		pg = append(pg, p)
	}

	// List available pages from devtool
	targets, err := b.devtool.List(ctx)
	if err != nil {
		return nil, err
	}

	// Add new targets to the list
	for _, t := range targets {
		if t.Type != devtool.Page {
			continue
		}

		var present bool
		for _, p := range pg {
			if t.ID == p.(*page).id {
				present = true
				break
			}
		}

		if !present {
			p, err := newPage(ctx, t, b.logger)
			if err != nil {
				return nil, err
			}
			pg = append(pg, p)
		}
	}

	// Update internal pages list
	b.pages = pg

	return &BrowserGetPagesOutput{Pages: pg}, nil
}

// Close shuts down the browser and cleans up resources.
func (b *browser) Close(ctx context.Context) error {
	b.logger.Debug("closing pages", "len", len(b.pages))

	b.mux.RLock()
	for _, p := range b.pages {
		p := p.(*page)
		if p.closed {
			b.logger.Debug("page already closed", "target_id", p.target.ID)
			continue
		}
		b.logger.Debug("closing page", "target_id", p.target.ID)
		err := p.Close(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}
	b.mux.RUnlock()

	b.devtool = nil
	b.pages = nil

	err := b.instance.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	defer func() {
		if _, err = os.Stat(b.datadir); !os.IsNotExist(err) {
			err = os.RemoveAll(b.datadir)
			if err != nil {
				b.logger.Warn("removing data dir error", "path", b.datadir, "error", err)
			} else {
				b.logger.Debug("removed data dir", "path", b.datadir)
			}
		}
	}()

	return <-b.waitChan
}

// GetDevToolClient retrieves the DevTools client associated with the browser.
// This client allows for advanced interactions with the browser's DevTools protocol,
// enabling custom actions and low-level debugging or profiling features.
func (b *browser) GetDevToolClient() *devtool.DevTools {
	return b.devtool
}
