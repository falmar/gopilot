package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/network"
	cdppage "github.com/mafredri/cdp/protocol/page"
)

type PageNavigation interface {
	// Activate brings page to front
	Activate(ctx context.Context) error

	// Navigate navigates the page to the specified URL.
	// The input is a PageNavigateInput containing the URL to navigate to.
	// It returns a PageNavigateOutput or an error if the navigation fails.
	Navigate(ctx context.Context, in *PageNavigateInput) (*PageNavigateOutput, error)

	// Reload reloads the current page.
	// It can take a PageReloadInput and returns a PageReloadOutput or an error.
	Reload(ctx context.Context, in *PageReloadInput) (*PageReloadOutput, error)
}

// PageNavigateInput specifies the input for the Navigate method.
// URL is the target URL to navigate to.
// WaitDomContentLoad determines whether to wait for the DOM content to load.
type PageNavigateInput struct {
	URL                string // The URL to navigate to.
	WaitDomContentLoad bool   // If true, waits for the DOM content to load before returning.
}

// PageNavigateOutput represents the output of the Navigate method.
// LoaderID is the ID associated with the loading process of the page.
type PageNavigateOutput struct {
	LoaderID network.LoaderID // The LoaderID associated with the navigation.
}

// Navigate navigates the page to the specified URL.
// Based on the input, it optionally waits for the DOM content to finish loading.
// Returns a PageNavigateOutput containing the LoaderID or an error if navigation fails.
func (p *page) Navigate(ctx context.Context, in *PageNavigateInput) (*PageNavigateOutput, error) {
	var domEvent cdppage.DOMContentEventFiredClient
	var err error

	if in.WaitDomContentLoad {
		// Open a DOMContentEventFired client to buffer this event.
		domEvent, err = p.client.Page.DOMContentEventFired(ctx)
		if err != nil {
			return nil, err
		}
		defer domEvent.Close()
	}

	p.logger.Debug("page navigation started", "url", in.URL)

	rp, err := p.client.Page.Navigate(ctx, &cdppage.NavigateArgs{URL: in.URL})
	if err != nil {
		return nil, err
	}

	if in.WaitDomContentLoad && domEvent != nil {
		p.logger.Debug("page waiting for DOM content to load")
		// Wait for the DOMContentEventFired
		if _, err = domEvent.Recv(); err != nil {
			return nil, err
		}
	}

	p.logger.Debug("page navigation finished", "frame", rp.FrameID)

	var loaderId network.LoaderID
	if rp.LoaderID != nil {
		loaderId = *rp.LoaderID
	}

	return &PageNavigateOutput{LoaderID: loaderId}, nil
}

// PageReloadInput specifies the input for the Reload method.
// LoaderID is the ID associated with the previous loading process.
// WaitDomContentLoad determines whether to wait for the DOM content to load after reloading.
type PageReloadInput struct {
	LoaderID           network.LoaderID // The LoaderID of the previous load.
	WaitDomContentLoad bool             // If true, waits for the DOM content to load after reload.
}

// PageReloadOutput represents the output of the Reload method.
type PageReloadOutput struct{}

// Reload reloads the current page.
// It optionally waits for the DOM content to finish loading before returning.
// Returns a PageReloadOutput or an error if reload fails.
func (p *page) Reload(ctx context.Context, in *PageReloadInput) (*PageReloadOutput, error) {
	var domEvent cdppage.DOMContentEventFiredClient
	var err error

	if in.WaitDomContentLoad {
		// Open a DOMContentEventFired client to buffer this event.
		domEvent, err = p.client.Page.DOMContentEventFired(ctx)
		if err != nil {
			return nil, err
		}
		defer domEvent.Close()
	}

	logger := p.logger
	if in.LoaderID != "" {
		logger = p.logger.With("loader_id", in.LoaderID) // Log with loader ID context
	}

	logger.Debug("reloading page")

	args := &cdppage.ReloadArgs{}
	if in.LoaderID != "" {
		args.LoaderID = &in.LoaderID
	}

	if err = p.client.Page.Reload(ctx, args); err != nil {
		return nil, err
	}

	if in.WaitDomContentLoad && domEvent != nil {
		p.logger.Debug("page waiting for DOM content to load")
		_, err = domEvent.Recv() // Wait for the DOMContentEventFired
		if err != nil {
			return nil, err
		}
	}

	return &PageReloadOutput{}, nil
}
