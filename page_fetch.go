package gopilot

import (
	"context"
	"errors"
	"fmt"

	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
)

type PageFetch interface {
	// EnableFetch enables network fetch interception.
	// Returns an error if enabling fails.
	EnableFetch(ctx context.Context) error

	// DisableFetch disables network fetch interception.
	// Returns an error if disabling fails.
	DisableFetch(ctx context.Context) error

	// AddInterceptRequest adds a request interception callback.
	// Returns a handle that can be used to remove the callback later.
	AddInterceptRequest(ctx context.Context, cb InterceptRequestCallback) *InterceptRequestHandle

	// RemoveInterceptRequest removes a request interception callback.
	// The handle parameter should be the value returned by AddInterceptRequest.
	RemoveInterceptRequest(ctx context.Context, handle *InterceptRequestHandle)

	// AddInterceptResponse adds a response interception callback.
	// Returns a handle that can be used to remove the callback later.
	AddInterceptResponse(ctx context.Context, cb InterceptResponseCallback) *InterceptResponseHandle

	// RemoveInterceptResponse removes a response interception callback.
	// The handle parameter should be the value returned by AddInterceptResponse.
	RemoveInterceptResponse(ctx context.Context, handle *InterceptResponseHandle)
}

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
	return p.handleRequestPaused(ctx)
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
// The callback receives details about the paused request and can modify it or provide a custom response.
// Return values:
// - (nil, nil): Continue the request with any modifications made to continueArgs
// - (nil, error): Abort the request with the given error
// - (*fetch.FulfillRequestArgs, nil): Fulfill the request with a custom response
type InterceptRequestCallback func(ctx context.Context, req *fetch.RequestPausedReply, continueArgs *fetch.ContinueRequestArgs) (*fetch.FulfillRequestArgs, error)

// InterceptRequestHandle is a handle for managing request interception callbacks.
type InterceptRequestHandle struct {
	cb InterceptRequestCallback
}

// AddInterceptRequest adds a request interception callback.
// Returns a handle to manage the interception callback.
func (p *page) AddInterceptRequest(_ context.Context, cb InterceptRequestCallback) *InterceptRequestHandle {
	p.mux.Lock()
	handle := &InterceptRequestHandle{}
	handle.cb = cb
	p.interceptRequests[handle] = cb
	p.mux.Unlock()
	return handle
}

// RemoveInterceptRequest removes a request interception callback using the provided handle.
// The callback associated with the handle is deleted.
func (p *page) RemoveInterceptRequest(_ context.Context, handle *InterceptRequestHandle) {
	p.mux.Lock()
	delete(p.interceptRequests, handle)
	handle.cb = nil
	p.mux.Unlock()
}

// InterceptResponseCallback is a function type for response interception.
// The callback receives details about the paused response and can modify it.
// If an error is returned, the response processing will be interrupted.
type InterceptResponseCallback func(ctx context.Context, req *fetch.RequestPausedReply, continueArgs *fetch.ContinueResponseArgs) error

// InterceptResponseHandle is a handle for managing response interception callbacks.
type InterceptResponseHandle struct {
	cb InterceptResponseCallback
}

// AddInterceptResponse adds a response interception callback.
// Returns a handle to manage the interception callback.
func (p *page) AddInterceptResponse(_ context.Context, cb InterceptResponseCallback) *InterceptResponseHandle {
	p.mux.Lock()
	handle := &InterceptResponseHandle{}
	handle.cb = cb
	p.interceptResponses[handle] = cb
	p.mux.Unlock()
	return handle
}

// RemoveInterceptResponse removes a response interception callback using the provided handle.
// The callback associated with the handle is deleted.
func (p *page) RemoveInterceptResponse(_ context.Context, handle *InterceptResponseHandle) {
	p.mux.Lock()
	delete(p.interceptResponses, handle)
	handle.cb = nil
	p.mux.Unlock()
}

// handleRequestPaused manages the received paused requests and responses,
// invoking the respective callbacks for each paused request or response.
func (p *page) handleRequestPaused(ctx context.Context) error {
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
			var continueRequest *fetch.ContinueRequestArgs
			var continueResponse *fetch.ContinueResponseArgs
			var fulfillRequest *fetch.FulfillRequestArgs

			p.mux.RLock()
			if isResponse {
				continueResponse = &fetch.ContinueResponseArgs{RequestID: rp.RequestID}
				for _, cb := range p.interceptResponses {
					callbackErr = cb(ctx, rp, continueResponse)
					if callbackErr != nil {
						break
					}
				}
			} else {
				continueRequest = &fetch.ContinueRequestArgs{RequestID: rp.RequestID}
				for _, cb := range p.interceptRequests {
					var err error
					fulfillRequest, err = cb(ctx, rp, continueRequest)
					if err != nil {
						callbackErr = err
						break
					}
					if fulfillRequest != nil {
						break
					}
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

			if !isResponse && fulfillRequest != nil {
				// Fulfill the request with custom data
				fulfillRequest.RequestID = rp.RequestID
				if err := p.client.Fetch.FulfillRequest(ctx, fulfillRequest); err != nil {
					p.logger.Warn("unable to fulfill request", "error", err, "url", rp.Request.URL)
				}
				continue
			}

			// set RequestID just in case
			if isResponse {
				continueResponse.RequestID = rp.RequestID
				callbackErr = p.client.Fetch.ContinueResponse(ctx, continueResponse)
			} else {
				continueRequest.RequestID = rp.RequestID
				callbackErr = p.client.Fetch.ContinueRequest(ctx, continueRequest)
			}

			if callbackErr != nil {
				p.logger.Warn("unable to continue request/response", "error", callbackErr, "url", rp.Request.URL)
			}
		}
	}()

	return nil
}
