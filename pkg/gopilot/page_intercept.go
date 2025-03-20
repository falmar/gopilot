package gopilot

import (
	"context"
	"errors"
	"fmt"

	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
)

// EnableFetch enables network request interception.
// It sets up the fetching mechanism and allows handling of authentication requests.
// Returns an error if enabling fails.
func (p *page) EnableFetch(ctx context.Context) error {
	if p.fetchEnabled {
		return nil
	}

	auth := true
	pattern := "*"
	enableArg := &fetch.EnableArgs{
		HandleAuthRequests: &auth,
		Patterns: []fetch.RequestPattern{
			{RequestStage: fetch.RequestStageNotSet, URLPattern: &pattern},
			{RequestStage: fetch.RequestStageRequest, URLPattern: &pattern},
			{RequestStage: fetch.RequestStageResponse, URLPattern: &pattern},
		},
	}

	err := p.client.Fetch.Enable(ctx, enableArg)
	if err != nil {
		return err
	}
	p.fetchEnabled = true
	return p.handleInterceptRequest(ctx)
}

// DisableFetch disables network request interception.
// Returns an error if disabling fails.
func (p *page) DisableFetch(ctx context.Context) error {
	if p.fetchEnabled {
		if err := p.interceptClient.Close(); err != nil {
			p.logger.Debug("unable to close paused request handler", "error", err)
		}
	}

	if err := p.client.Fetch.Disable(ctx); err != nil {
		return fmt.Errorf("unable to disable fetch: %w", err)
	}
	p.fetchEnabled = false
	return nil
}

// InterceptRequestCallback is a function type for request interception.
// It allows users to define logic for handling intercepted requests.
// The callback receives the current context and a RequestPausedReply,
// which includes details about the paused request.
// If an error is returned, the request will be aborted,
// providing flexibility to control network requests during automation tasks.
type InterceptRequestCallback func(ctx context.Context, req *fetch.RequestPausedReply) error

// InterceptRequestHandle is a handle for managing request interception callbacks.
type InterceptRequestHandle struct{}

// AddInterceptRequest adds a request interception callback.
// It returns a handle to manage the interception callback.
func (p *page) AddInterceptRequest(_ context.Context, cb InterceptRequestCallback) *InterceptRequestHandle {
	p.mux.Lock()
	handle := &InterceptRequestHandle{}
	p.interceptRequests[handle] = cb
	p.mux.Unlock()
	return handle
}

// RemoveInterceptRequest removes a request interception callback using the provided handle.
// The callback associated with the handle is deleted.
func (p *page) RemoveInterceptRequest(_ context.Context, handle *InterceptRequestHandle) {
	p.mux.Lock()
	delete(p.interceptRequests, handle)
	p.mux.Unlock()
}

// handleInterceptRequest manages the received paused requests,
// invoking the respective callbacks for each paused request.
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

			p.logger.Debug("received paused request", "request_id", rp.RequestID, "url", rp.Request.URL, "resource_type", rp.ResourceType, "response", isResponse)

			var callbackErr error
			p.mux.RLock()
			for _, cb := range p.interceptRequests {
				callbackErr = cb(ctx, rp)
				if callbackErr != nil {
					break
				}
			}
			p.mux.RUnlock()

			if callbackErr != nil {
				if !isResponse {
					if err := p.client.Fetch.FailRequest(ctx, &fetch.FailRequestArgs{
						RequestID:   rp.RequestID,
						ErrorReason: network.ErrorReasonAborted,
					}); err != nil {
						p.logger.Warn("unable to abort request", "error", err, "url", rp.Request.URL)
					}
				}
				continue
			}

			if isResponse {
				callbackErr = p.client.Fetch.ContinueResponse(ctx, &fetch.ContinueResponseArgs{
					RequestID: rp.RequestID,
				})
			} else {
				callbackErr = p.client.Fetch.ContinueRequest(ctx, &fetch.ContinueRequestArgs{
					RequestID: rp.RequestID,
				})
			}

			if callbackErr != nil {
				p.logger.Warn("unable to continue request/response", "error", callbackErr, "url", rp.Request.URL)
			}
		}
	}()

	return nil
}
