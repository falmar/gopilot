package gopilot

import (
	"context"
	"log/slog"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"

	cdpdevtool "github.com/mafredri/cdp/devtool"
	cdppage "github.com/mafredri/cdp/protocol/page"
)

type InterceptRequestCallback func(ctx context.Context, req *fetch.RequestPausedReply) error
type InterceptRequestHandle struct{}

type Page interface {
	Navigate(ctx context.Context, in *PageNavigateInput) (*PageNavigateOutput, error)
	Reload(ctx context.Context, in *PageReloadInput) (*PageReloadOutput, error)
	GetContent(ctx context.Context) (string, error)
	Close(ctx context.Context) error

	EnableFetch(ctx context.Context) error
	DisableFetch(ctx context.Context) error
	AddInterceptRequest(ctx context.Context, cb InterceptRequestCallback) *InterceptRequestHandle
	RemoveInterceptRequest(ctx context.Context, handle *InterceptRequestHandle)

	Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error)
	QuerySelector(ctx context.Context, in *PageQuerySelectorInput) (*PageQuerySelectorOutput, error)

	GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error)
	SetCookies(ctx context.Context, in *SetCookiesInput) (*SetCookiesOutput, error)
	ClearCookies(ctx context.Context, in *ClearCookiesInput) (*ClearCookiesOutput, error)
}

type page struct {
	devtool *devtool.DevTools
	target  *devtool.Target
	conn    *rpcc.Conn
	client  *cdp.Client
	logger  *slog.Logger

	mux    sync.RWMutex
	closed bool

	domEvent cdppage.DOMContentEventFiredClient

	fetchEnabled      bool
	interceptClient   fetch.RequestPausedClient
	interceptRequests map[*InterceptRequestHandle]InterceptRequestCallback
}

func newPage(
	ctx context.Context,
	devtool *devtool.DevTools,
	logger *slog.Logger,
	newTab bool,
) (Page, error) {
	logger.Debug("creating new page cdp target")

	var target *cdpdevtool.Target
	var err error = nil

	if newTab {
		target, err = devtool.Create(ctx)
	} else {
		target, err = devtool.Get(ctx, cdpdevtool.Page)
	}
	if err != nil {
		return nil, err
	}

	logger.Debug("creating rpc conn")
	conn, err := rpcc.DialContext(ctx, target.WebSocketDebuggerURL)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating protocol client")
	client := cdp.NewClient(conn)

	p := &page{
		devtool: devtool,
		client:  client,
		target:  target,
		conn:    conn,
		logger:  logger,
		mux:     sync.RWMutex{},

		interceptRequests: map[*InterceptRequestHandle]InterceptRequestCallback{},
	}

	// Enable events on the Page domain, it's often preferrable to create
	// event clients before enabling events so that we don't miss any.
	if err = p.client.Page.Enable(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

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

type PageEvaluateInput struct {
	AwaitPromise bool
	ReturnValue  bool
	Expression   string
}
type PageEvaluateOutput struct {
	Value []byte
}

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
