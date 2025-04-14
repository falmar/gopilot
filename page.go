package gopilot

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
)

// Page represents a web page in the browser.
type Page interface {
	PageNavigation
	PageDOM
	PageFetch
	PageStorage

	// Close closes the page.
	// Returns an error if closing the page fails.
	Close(ctx context.Context) error

	// Evaluate runs JavaScript on the page.
	// Takes a PageEvaluateInput and returns a PageEvaluateOutput or an error.
	Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error)

	// TakeScreenshot captures a screenshot of the page.
	// You can choose to capture the entire page or just the visible viewport.
	// Input parameters allow you to specify the image format and capture area.
	// Returns the screenshot data or an error if the capture fails.
	TakeScreenshot(ctx context.Context, in *PageTakeScreenshotInput) (*PageTakeScreenshotOutput, error)

	// GetTargetID returns the unique identifier for the page's target.
	// This ID can be used to distinguish different pages or targets in the browser.
	GetTargetID() string

	// GetCDPClient retrieves the Chrome DevTools Protocol (CDP) client associated with the page.
	// The CDP client allows for direct communication with the browser's protocol.
	// This is useful for performing low-level operations and custom actions not exposed by higher-level methods.
	GetCDPClient() *cdp.Client
}

type page struct {
	id     string
	target *devtool.Target
	conn   *rpcc.Conn
	client *cdp.Client
	logger *slog.Logger
	mux    sync.RWMutex
	closed bool

	fetchEnabled      bool
	interceptClient   fetch.RequestPausedClient
	interceptRequests map[*InterceptRequestHandle]InterceptRequestCallback
}

// newPage creates a new Page instance.
// It initializes connection and protocol client, and enables page events.
func newPage(
	ctx context.Context,
	t *devtool.Target,
	logger *slog.Logger,
) (Page, error) {
	logger.Debug("creating rpc conn")
	conn, err := rpcc.DialContext(ctx, t.WebSocketDebuggerURL,
		rpcc.WithWriteBufferSize(1024*1024*5), // 5mb
		rpcc.WithCompression())
	if err != nil {
		return nil, err
	}

	logger.Debug("creating protocol client")
	client := cdp.NewClient(conn)
	p := &page{
		id:                t.ID,
		client:            client,
		target:            t,
		conn:              conn,
		logger:            logger,
		mux:               sync.RWMutex{},
		interceptRequests: map[*InterceptRequestHandle]InterceptRequestCallback{},
	}

	// Enable events on the Page domain, it's often preferable to create
	// event clients before enabling events so that we don't miss any.
	if err = p.client.Page.Enable(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func getPageCurrentURL(ctx context.Context, p *page) (*url.URL, error) {
	// TODO: listen for target url changes

	rValue := true
	rp, err := p.client.Runtime.Evaluate(ctx, &runtime.EvaluateArgs{
		Expression:    `window.location.toString()`,
		ReturnByValue: &rValue,
	})
	if err != nil {
		return nil, err
	}

	var pageURL string
	err = json.Unmarshal(rp.Result.Value, &pageURL)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(pageURL)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Close closes the page and underlying connections.
func (p *page) Close(ctx context.Context) error {
	defer p.conn.Close()

	err := p.client.Page.Close(ctx)
	if err != nil {
		return err
	}

	p.mux.Lock()
	p.closed = true
	p.mux.Unlock()

	return nil
}

func (p *page) Activate(ctx context.Context) error {
	return p.client.Page.BringToFront(ctx)
}

// GetTargetID returns the unique identifier for the page's target.
// This ID can be used to distinguish different pages or targets in the browser.
func (p *page) GetTargetID() string {
	return p.id
}

// GetCDPClient retrieves the Chrome DevTools Protocol (CDP) client associated with the page.
// The CDP client allows for direct communication with the browser's protocol.
// This is useful for performing low-level operations and custom actions not exposed by higher-level methods.
func (p *page) GetCDPClient() *cdp.Client {
	return p.client
}
