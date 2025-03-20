package gopilot

import (
	"context"
	"strings"

	"github.com/mafredri/cdp/protocol/fetch"
	"github.com/mafredri/cdp/protocol/network"
)

// XHREvent represents an XHR event with related information.
type XHREvent struct {
	URL    string // The URL that was requested
	Body   string // The body of the response
	Base64 bool   // Indicates if the response body is Base64 encoded
	Error  error  // Error encountered during the request (if any)
}

// XHRMonitor is an interface for monitoring XHR requests.
type XHRMonitor interface {
	// Listen starts listening for XHR events that match the provided patterns.
	// It returns a channel of XHREvent and an error if the operation fails.
	Listen(ctx context.Context, patterns []string) (chan *XHREvent, error)

	// Stop stops monitoring the XHR requests.
	// Returns an error if stopping fails.
	Stop(ctx context.Context) error
}

// NewXHRMonitor creates a new XHRMonitor instance.
// It takes a Page and returns an instance of XHRMonitor.
func NewXHRMonitor(p Page) XHRMonitor {
	return &xhrMonitor{
		p: p,
		c: make(chan *XHREvent, 100),
	}
}

// xhrMonitor is a concrete implementation of the XHRMonitor interface.
type xhrMonitor struct {
	p        Page                    // Associated Page
	c        chan *XHREvent          // Channel for XHREvents
	cbHandle *InterceptRequestHandle // Handle to the request interception
}

// Listen starts listening for XHR events that match the given patterns.
func (m *xhrMonitor) Listen(ctx context.Context, patterns []string) (chan *XHREvent, error) {
	p := m.p.(*page)

	// Enable fetch interception for capturing network requests.
	err := p.EnableFetch(ctx)
	if err != nil {
		return nil, err
	}

	hasPatterns := len(patterns) > 0
	m.cbHandle = p.AddInterceptRequest(ctx, InterceptRequestCallback(func(ctx context.Context, rp *fetch.RequestPausedReply) error {
		// Filter out non-XHR and non-fetch requests.
		if rp.ResourceType != network.ResourceTypeFetch && rp.ResourceType != network.ResourceTypeXHR {
			return nil
		}

		isResponse := rp.ResponseStatusCode != nil && *rp.ResponseStatusCode > 0
		if !isResponse {
			return nil
		}

		// Check if the request matches any given patterns.
		match := true
		if hasPatterns {
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

		// Fetch the response body.
		rb, err := p.client.Fetch.GetResponseBody(ctx, &fetch.GetResponseBodyArgs{RequestID: rp.RequestID})
		ev := &XHREvent{
			URL:   rp.Request.URL,
			Error: err,
		}

		// Set the body and encoding status if there were no errors.
		if err == nil {
			ev.Body = rb.Body
			ev.Base64 = rb.Base64Encoded
		}

		// Send the event to the channel.
		m.c <- ev
		return nil
	}))

	return m.c, nil
}

// Stop stops monitoring XHR requests.
func (m *xhrMonitor) Stop(ctx context.Context) error {
	p := m.p.(*page)
	p.RemoveInterceptRequest(ctx, m.cbHandle)
	return nil
}
