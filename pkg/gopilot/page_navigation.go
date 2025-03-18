package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/network"
	cdppage "github.com/mafredri/cdp/protocol/page"
)

type PageNavigateInput struct {
	URL                string
	WaitDomContentLoad bool
}
type PageNavigateOutput struct {
	LoaderID network.LoaderID
}

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

	p.logger.Debug("page navigating", "url", in.URL)
	rp, err := p.client.Page.Navigate(ctx, &cdppage.NavigateArgs{URL: in.URL})
	if err != nil {
		return nil, err
	}

	if in.WaitDomContentLoad && domEvent != nil {
		p.logger.Debug("page waiting dom content load")
		_, err = domEvent.Recv()
		if err != nil {
			return nil, err
		}
	}

	p.logger.Debug("navigated", "frame", rp.FrameID)

	var loaderId network.LoaderID
	if rp.LoaderID != nil {
		loaderId = *rp.LoaderID
	}

	return &PageNavigateOutput{LoaderID: loaderId}, nil
}

type PageReloadInput struct {
	LoaderID           network.LoaderID
	WaitDomContentLoad bool
}
type PageReloadOutput struct{}

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
		logger = p.logger.With("loader_id", in.LoaderID)
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
		p.logger.Debug("page waiting dom content load")
		_, err = domEvent.Recv()
		if err != nil {
			return nil, err
		}
	}

	return &PageReloadOutput{}, nil
}
