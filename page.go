package gopilot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/devtool"
	"github.com/mafredri/cdp/protocol/dom"
	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/runtime"
	"github.com/mafredri/cdp/rpcc"

	cdpdevtool "github.com/mafredri/cdp/devtool"
	cdppage "github.com/mafredri/cdp/protocol/page"
)

type InterceptRequestCallback func(ctx context.Context, req *fetch.RequestPausedReply) error
type InterceptRequestHandle struct{}

type Page interface {
	Navigate(ctx context.Context, url string) error
	GetContent(ctx context.Context) (string, error)
	Close(ctx context.Context) error

	EnableFetch(ctx context.Context) error
	DisableFetch(ctx context.Context) error
	AddInterceptRequest(ctx context.Context, cb InterceptRequestCallback) *InterceptRequestHandle
	RemoveInterceptRequest(ctx context.Context, handle *InterceptRequestHandle)

	Evaluate(ctx context.Context, in *PageEvaluateInput) (*PageEvaluateOutput, error)
	QuerySelector(ctx context.Context, query string) (interface{}, error)

	//ListenXHR(ctx context.Context, patterns []string) (chan *XHREvent, error)
}

type page struct {
	devtool *devtool.DevTools
	target  *devtool.Target
	conn    *rpcc.Conn
	client  *cdp.Client
	logger  *slog.Logger

	mux sync.RWMutex

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

func (p *page) Navigate(ctx context.Context, url string) error {
	// Open a DOMContentEventFired client to buffer this event.
	domEvent, err := p.client.Page.DOMContentEventFired(ctx)
	if err != nil {
		return err
	}
	defer domEvent.Close()

	p.logger.Debug("page navigating", "url", url)
	rp, err := p.client.Page.Navigate(ctx, cdppage.NewNavigateArgs(url))
	if err != nil {
		return err
	}

	p.logger.Debug("page waiting dom content load")
	_, err = domEvent.Recv()
	if err != nil {
		return err
	}

	p.logger.Debug("navigated", "frame", rp.FrameID)

	return nil
}

func (p *page) Close(ctx context.Context) error {
	defer p.conn.Close()

	return p.client.Page.Close(ctx)
}

func (p *page) GetContent(ctx context.Context) (string, error) {
	doc, err := p.client.DOM.GetDocument(ctx, nil)
	if err != nil {
		return "", err
	}
	rp, err := p.client.DOM.GetOuterHTML(ctx, &dom.GetOuterHTMLArgs{
		NodeID: &doc.Root.NodeID,
	})
	if err != nil {
		return "", err
	}

	return rp.OuterHTML, nil
}

func (p *page) EnableFetch(ctx context.Context) error {
	if p.fetchEnabled {
		return nil
	}

	auth := true
	enableArg := &fetch.EnableArgs{
		HandleAuthRequests: &auth,
	}

	pattern := "*"
	enableArg.Patterns = append(enableArg.Patterns, fetch.RequestPattern{
		RequestStage: fetch.RequestStageNotSet,
		URLPattern:   &pattern,
	})
	enableArg.Patterns = append(enableArg.Patterns, fetch.RequestPattern{
		RequestStage: fetch.RequestStageRequest,
		URLPattern:   &pattern,
	})
	enableArg.Patterns = append(enableArg.Patterns, fetch.RequestPattern{
		RequestStage: fetch.RequestStageResponse,
		URLPattern:   &pattern,
	})

	err := p.client.Fetch.Enable(ctx, enableArg)
	if err != nil {
		return err
	}

	p.fetchEnabled = true

	return p.handleInterceptRequest(ctx)
}

func (p *page) DisableFetch(ctx context.Context) error {
	if p.fetchEnabled {
		if err := p.interceptClient.Close(); err != nil {
			p.logger.Debug("unable to close paused request handler", "error", err)
		}
	}

	if err := p.client.Fetch.Disable(ctx); err != nil {
		return fmt.Errorf("unable to disable fetch %w", err)
	}

	p.fetchEnabled = false

	return nil
}

func (p *page) AddInterceptRequest(_ context.Context, cb InterceptRequestCallback) *InterceptRequestHandle {
	p.mux.Lock()
	handle := &InterceptRequestHandle{}
	p.interceptRequests[handle] = cb
	p.mux.Unlock()
	return handle
}

func (p *page) RemoveInterceptRequest(_ context.Context, handle *InterceptRequestHandle) {
	p.mux.Lock()
	delete(p.interceptRequests, handle)
	p.mux.Unlock()
}

func (p *page) handleInterceptRequest(ctx context.Context) error {

	if err := p.EnableFetch(ctx); err != nil {
		return err
	}

	pc, err := p.client.Fetch.RequestPaused(ctx)
	if err != nil {
		return err
	}

	p.interceptClient = pc

	go func() {
		defer pc.Close()

		for {
			rp, err := pc.Recv()
			if err != nil && !errors.Is(err, context.DeadlineExceeded) {
				return
			} else if errors.Is(err, context.DeadlineExceeded) {
				return
			}

			isResponse := rp.ResponseStatusCode != nil && *rp.ResponseStatusCode > 0

			p.logger.Debug("received paused request",
				"request_id", rp.RequestID,
				"url", rp.Request.URL,
				"resource_type", rp.ResourceType,
				"response", isResponse,
			)

			err = nil

			p.mux.RLock()
			for _, cb := range p.interceptRequests {
				err = cb(ctx, rp)
				if err != nil {
					break
				}
			}
			p.mux.RUnlock()

			if err != nil {
				if !isResponse {
					err = p.client.Fetch.FailRequest(ctx, &fetch.FailRequestArgs{
						RequestID:   rp.RequestID,
						ErrorReason: network.ErrorReasonAborted,
					})

					if err != nil {
						p.logger.Warn("unable to abort request/response", "error", err, "url", rp.Request.URL)
					}
				}

				continue
			}

			if isResponse {
				err = p.client.Fetch.ContinueResponse(ctx, &fetch.ContinueResponseArgs{RequestID: rp.RequestID})
			} else {
				err = p.client.Fetch.ContinueRequest(ctx, &fetch.ContinueRequestArgs{RequestID: rp.RequestID})
			}

			if err != nil {
				p.logger.Warn("unable to continue request/response", "error", err, "url", rp.Request.URL)
			}
		}
	}()

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

func (p *page) QuerySelector(ctx context.Context, query string) (interface{}, error) {
	return nil, errors.New("not implemented")
}
