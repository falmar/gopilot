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

	// GetPages retrieves only pages created by this session (tracked pages).
	// These are pages created with NewTab: true. Calling Close() on these pages will close them.
	// Returns a BrowserGetPagesOutput with a list of session pages or an error if retrieving fails.
	GetPages(ctx context.Context, in *BrowserGetPagesInput) (*BrowserGetPagesOutput, error)

	// GetAllPages retrieves ALL pages in the browser, including non-session pages.
	// Pages returned are NOT session-tracked, and calling Close() on them is a no-op.
	// Use this for inspection/debugging. For pages created by this session, use GetPages().
	GetAllPages(ctx context.Context, in *BrowserGetPagesInput) (*BrowserGetPagesOutput, error)

	// Close shuts down the browser instance and cleans up any resources.
	// Only closes session pages (pages created by this instance with NewTab: true).
	// For external browsers, closes session pages but leaves the browser running.
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
	waitChan chan error

	isExternal   bool             // true if connected to external browser via ConnectionURL
	sessionPages map[string]*page // tracks pages created by this session (target ID -> page)
}

// NewBrowser creates a new browser instance with the given configuration and logger.
func NewBrowser(cfg *BrowserConfig, logger *slog.Logger) Browser {
	return &browser{
		config:       cfg,
		logger:       logger,
		waitChan:     make(chan error, 1),
		sessionPages: make(map[string]*page),
	}
}

func (b *browser) cleanup() {
	if b.instance != nil && b.instance.Process != nil {
		if err := b.instance.Process.Kill(); err != nil {
			b.logger.Debug("kill browser error", "error", err)
		}
		<-b.waitChan
	}
	if b.datadir != "" {
		if err := os.RemoveAll(b.datadir); err != nil {
			b.logger.Warn("remove data dir error", "path", b.datadir, "error", err)
		}
		b.datadir = ""
	}
}

func (b *browser) addSessionPage(p *page) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.sessionPages[p.id] = p
}

func (b *browser) removeSessionPage(targetID string) {
	b.mux.Lock()
	defer b.mux.Unlock()
	delete(b.sessionPages, targetID)
}

func (b *browser) isSessionPage(targetID string) bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	_, exists := b.sessionPages[targetID]
	return exists
}

func (b *browser) getSessionPagesSnapshot() []*page {
	b.mux.RLock()
	defer b.mux.RUnlock()

	pages := make([]*page, 0, len(b.sessionPages))
	for _, p := range b.sessionPages {
		pages = append(pages, p)
	}
	return pages
}

func (b *browser) clearSessionPages() {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.sessionPages = make(map[string]*page)
}

// BrowserOpenInput contains parameters required to open a browser.
type BrowserOpenInput struct{}

// Open initializes and starts the browser process, or connects to an existing browser.
func (b *browser) Open(ctx context.Context, in *BrowserOpenInput) (err error) {
	if b.config.ConnectionURL != "" {
		if err := b.openExternal(ctx); err != nil {
			b.clearSessionPages()
			return err
		}
		return nil
	}

	// Original startup flow for launching a new browser process
	tempDir, err := os.MkdirTemp("", "gopilot")
	if err != nil {
		return err
	}
	b.datadir = tempDir
	b.logger.Debug("created data dir", "path", b.datadir)

	// ensure browser is close when fails early
	defer func() {
		if err != nil {
			b.logger.Info("open failed, cleaning up", "error", err)
			b.cleanup()
		}
	}()

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

	b.instance.Args = append(
		b.instance.Args,
		"about:blank",
	)

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

	if err = b.instance.Start(); err != nil {
		return err
	}
	b.logger.Debug("waiting for devtool url message")

	go func() {
		b.waitChan <- b.instance.Wait()
	}()

	waitDuration := defaultOpenTimeout
	if b.config.OpenTimeout != nil {
		waitDuration = *b.config.OpenTimeout
	}

	var devtoolsURLString string
	select {
	case <-ctx.Done():
		return ctx.Err()
	case werr := <-b.waitChan:
		b.waitChan <- werr
		return fmt.Errorf("exec wait exited unexpectedly or too soon: %w", werr)
	case <-time.After(waitDuration):
		return fmt.Errorf("duration %s exceeded waiting for devtool url", waitDuration)
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

	if _, err = b.devtool.Version(ctx); err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}

	return nil
}

// openExternal connects to an existing external browser instance.
func (b *browser) openExternal(ctx context.Context) error {
	b.isExternal = true

	parsedURL, err := url.Parse(b.config.ConnectionURL)
	if err != nil {
		return fmt.Errorf("invalid ConnectionURL: %w", err)
	}

	if parsedURL.Hostname() == "" {
		return errors.New("ConnectionURL must specify a valid hostname")
	}

	port := parsedURL.Port()
	if port == "" {
		return errors.New("ConnectionURL must specify a port")
	}

	// Support both WebSocket URLs (ws://host:port/devtools/browser/UUID) and HTTP (http://host:port)
	devtoolEndpoint := fmt.Sprintf("http://%s:%s", parsedURL.Hostname(), port)
	b.logger.Debug("connecting to external browser", "endpoint", devtoolEndpoint)
	b.devtool = devtool.New(devtoolEndpoint)

	// Validate connection by checking browser version
	if _, err := b.devtool.Version(ctx); err != nil {
		return fmt.Errorf("failed to connect to external browser at %s: %w", b.config.ConnectionURL, err)
	}

	b.logger.Debug("successfully connected to external browser", "endpoint", devtoolEndpoint)
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
// When NewTab is true, the page is tracked as a session page.
// When NewTab is false, the existing page is reused without tracking (not owned by session).
func (b *browser) NewPage(ctx context.Context, in *BrowserNewPageInput) (*BrowserNewPageOutput, error) {
	b.logger.Debug("creating new page cdp target")

	var t *devtool.Target
	var err error

	if in.NewTab {
		if in.URL == "" {
			t, err = b.devtool.CreateURL(ctx, "about:blank")
		} else {
			t, err = b.devtool.CreateURL(ctx, in.URL)
		}
	} else {
		t, err = b.devtool.Get(ctx, devtool.Page)
	}

	if err != nil {
		return nil, err
	}

	// Mark as session page only if we created it
	isSessionPage := in.NewTab

	p, err := newPage(ctx, t, b.logger, isSessionPage)
	if err != nil {
		return nil, err
	}

	// Track ONLY session pages
	if isSessionPage {
		pagePtr := p.(*page)
		pagePtr.browser = b
		b.addSessionPage(pagePtr)
		b.logger.Debug("created and tracked session page", "target_id", t.ID)
	} else {
		b.logger.Debug("reusing existing page (not tracked)", "target_id", t.ID)
	}

	return &BrowserNewPageOutput{Page: p}, nil
}

// BrowserGetPagesInput represents parameters to obtain open pages.
type BrowserGetPagesInput struct{}

// BrowserGetPagesOutput contains the list of open browser pages.
type BrowserGetPagesOutput struct {
	Pages []Page
}

// GetPages retrieves only the pages created by this session.
// These are pages created with NewTab: true and tracked by gopilot.
// For all pages in the browser (including non-session pages), use GetAllPages().
func (b *browser) GetPages(ctx context.Context, _ *BrowserGetPagesInput) (*BrowserGetPagesOutput, error) {
	pages := b.getSessionPagesSnapshot()

	pg := make([]Page, len(pages))
	for i, p := range pages {
		pg[i] = p
	}

	b.logger.Debug("retrieved session pages", "count", len(pg))
	return &BrowserGetPagesOutput{Pages: pg}, nil
}

// GetAllPages retrieves ALL pages in the browser, including non-session pages.
// Pages returned from this method are NOT session-tracked and calling Close() on them is a no-op.
// Use this for inspection/debugging. For pages created by this session, use GetPages().
func (b *browser) GetAllPages(ctx context.Context, _ *BrowserGetPagesInput) (*BrowserGetPagesOutput, error) {
	var pg []Page

	targets, err := b.getPageTargets(ctx)
	if err != nil {
		return nil, err
	}

	for _, t := range targets {
		// Mark as non-session page (isSessionPage: false)
		p, err := newPage(ctx, t, b.logger, false)
		if err != nil {
			return nil, err
		}
		pg = append(pg, p)
	}

	b.logger.Debug("retrieved all pages", "count", len(pg))
	return &BrowserGetPagesOutput{Pages: pg}, nil
}

// Close shuts down the browser and cleans up resources.
// Only closes session pages (pages created by this gopilot instance).
// For external browsers (ConnectionURL), closes session pages but leaves the browser running.
func (b *browser) Close(ctx context.Context) error {
	// allow a brief moment before closing pages
	time.Sleep(defaultClosePause)

	// Get snapshot of session pages to close
	pagesToClose := b.getSessionPagesSnapshot()
	b.clearSessionPages()

	// Close each session page via DevTools
	b.logger.Debug("closing session pages", "count", len(pagesToClose))
	for _, p := range pagesToClose {
		b.logger.Debug("closing session page", "target_id", p.id)

		err := b.devtool.Close(ctx, p.target)
		if err != nil && !errors.Is(err, context.Canceled) {
			b.logger.Debug("closing page error", "target_id", p.id, "err", err)
		}
	}

	b.devtool = nil

	// If connected to external browser, stop here (don't kill process or clean up temp dir)
	if b.isExternal {
		b.logger.Debug("external browser connection, skipping process cleanup")
		return nil
	}

	b.cleanup()
	return nil
}

// GetDevToolClient retrieves the DevTools client associated with the browser.
// This client allows for advanced interactions with the browser's DevTools protocol,
// enabling custom actions and low-level debugging or profiling features.
func (b *browser) GetDevToolClient() *devtool.DevTools {
	return b.devtool
}

func (b *browser) getPageTargets(ctx context.Context) ([]*devtool.Target, error) {
	// List available pages from devtool
	targets, err := b.devtool.List(ctx)
	if err != nil {
		return nil, err
	}

	// Add new targets to the list
	var pages []*devtool.Target
	for _, t := range targets {
		if t.Type != devtool.Page {
			continue
		}
		pages = append(pages, t)
	}
	return pages, nil
}
