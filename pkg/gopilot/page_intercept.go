package gopilot

import (
	"context"
	"errors"
	"fmt"

	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
)

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
