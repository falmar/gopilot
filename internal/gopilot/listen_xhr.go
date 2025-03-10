package gopilot

import (
	"context"
	"strings"

	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
)

type XHREvent struct {
	reqId  fetch.RequestID
	URL    string
	Body   string
	Base64 bool
	Error  error
}

type XHRMonitor interface {
	Listen(ctx context.Context, patterns []string) (chan *XHREvent, error)
	Stop(ctx context.Context) error
}

func NewXHRMonitor(p Page) XHRMonitor {
	return &xhrMonitor{
		p: p,
		c: make(chan *XHREvent, 100),
	}
}

type xhrMonitor struct {
	p        Page
	c        chan *XHREvent
	cbHandle *InterceptRequestHandle
}

func (m *xhrMonitor) Listen(ctx context.Context, patterns []string) (chan *XHREvent, error) {
	p := m.p.(*page)

	err := p.EnableFetch(ctx)
	if err != nil {
		return nil, err
	}

	hasPatters := len(patterns) > 0

	m.cbHandle = p.AddInterceptRequest(ctx, InterceptRequestCallback(func(ctx context.Context, rp *fetch.RequestPausedReply) error {
		if rp.ResourceType != network.ResourceTypeFetch && rp.ResourceType != network.ResourceTypeXHR {
			return nil
		}

		isResponse := rp.ResponseStatusCode != nil && *rp.ResponseStatusCode > 0
		if !isResponse {
			return nil
		}

		match := true

		if hasPatters {
			match = false

			for _, pattern := range patterns {
				if strings.Contains(rp.Request.URL, pattern) {
					match = true
					break
				}
			}
		}

		if !match {
			return nil
		}

		rb, err := p.client.Fetch.GetResponseBody(ctx, &fetch.GetResponseBodyArgs{
			RequestID: rp.RequestID,
		})

		ev := &XHREvent{
			URL:   rp.Request.URL,
			Error: err,
		}

		if err == nil {
			ev.Body = rb.Body
			ev.Base64 = rb.Base64Encoded
		}

		m.c <- ev

		return nil
	}))

	return m.c, nil
}

func (m *xhrMonitor) Stop(ctx context.Context) error {
	p := m.p.(*page)
	p.RemoveInterceptRequest(ctx, m.cbHandle)
	return nil
}
