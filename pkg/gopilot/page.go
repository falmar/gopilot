package gopilot

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"
)

// Page represents a web page in the browser.
type Page interface {
	// Navigate navigates the page to the specified URL.
	// The input is a PageNavigateInput containing the URL to navigate to.
	// It returns a PageNavigateOutput or an error if the navigation fails.
	Navigate(ctx context.Context, in *PageNavigateInput) (*PageNavigateOutput, error)

	// Reload reloads the current page.
	// It can take a PageReloadInput and returns a PageReloadOutput or an error.
	Reload(ctx context.Context, in *PageReloadInput) (*PageReloadOutput, error)

	// GetContent retrieves the HTML content of the page as a string.
	// Returns the content or an error if retrieving fails.
	GetContent(ctx context.Context) (string, error)

	// Close closes the page.
	// Returns an error if closing the page fails.
	Close(ctx context.Context) error

	// EnableFetch enables network fetch interception.
	// Returns an error if enabling fails.
	EnableFetch(ctx context.Context) error

	// DisableFetch disables network fetch interception.
	// Returns an error if disabling fails.
	DisableFetch(ctx context.Context) error

	// AddInterceptRequest adds a request interception callback.
	// It takes a callback function and returns an InterceptRequestHandle.
	AddInterceptRequest(ctx context.Context, cb InterceptRequestCallback) *InterceptRequestHandle

	// RemoveInterceptRequest removes a request interception callback.
	// It takes a handle to the callback to be removed.
	RemoveInterceptRequest(ctx context.Context, handle *InterceptRequestHandle)

	// Evaluate runs JavaScript on the page.
	// Takes a PageEvaluateInput and returns a PageEvaluateOutput or an error.
	Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error)

	// QuerySelector finds an element matching the selector.
	// Takes a PageQuerySelectorInput and returns a PageQuerySelectorOutput or an error.
	QuerySelector(ctx context.Context, in *PageQuerySelectorInput) (*PageQuerySelectorOutput, error)

	// Search finds a element matching the text, query selector or xpath
	// Takes a PageSearchInput and returns a PageSearchOutput or an error.
	Search(ctx context.Context, in *PageSearchInput) (*PageSearchOutput, error)

	// GetCookies retrieves cookies for the current page.
	// Takes a GetCookiesInput and returns GetCookiesOutput or an error.
	GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error)

	// SetCookies sets cookies for the current page.
	// Takes a SetCookiesInput and returns SetCookiesOutput or an error.
	SetCookies(ctx context.Context, in *SetCookiesInput) (*SetCookiesOutput, error)

	// ClearCookies clears cookies for the current page.
	// Takes a ClearCookiesInput and returns ClearCookiesOutput or an error.
	ClearCookies(ctx context.Context, in *ClearCookiesInput) (*ClearCookiesOutput, error)

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
	conn, err := rpcc.DialContext(ctx, t.WebSocketDebuggerURL)
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

// PageEvaluateInput specifies input for the Evaluate method.
type PageEvaluateInput struct {
	AwaitPromise bool
	ReturnValue  bool
	Expression   string
}

// PageEvaluateOutput represents the output of the Evaluate method.
type PageEvaluateOutput struct {
	Value json.RawMessage
}

// Evaluate executes the given JavaScript expression on the page.
func (p *page) Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error) {
	userGesture := true
	allowUnsafe := true

	res, err := p.client.Runtime.Evaluate(ctx, &runtime.EvaluateArgs{
		Expression:                  in.Expression,
		UserGesture:                 &userGesture,
		ReturnByValue:               &in.ReturnValue,
		AwaitPromise:                &in.AwaitPromise,
		AllowUnsafeEvalBlockedByCSP: &allowUnsafe,
	})
	if err != nil {
		return nil, err
	}

	out := &PageEvaluateOutput{}
	if in.ReturnValue {
		out.Value = res.Result.Value
	}

	return out, nil
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
