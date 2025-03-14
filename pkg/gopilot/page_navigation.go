package gopilot

import (
	"context"

	cdppage "github.com/mafredri/cdp/protocol/page"
)

type PageNavigateInput struct {
	URL                string
	WaitDomContentLoad bool
}
type PageNavigateOutput struct{}

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
	rp, err := p.client.Page.Navigate(ctx, cdppage.NewNavigateArgs(in.URL))
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

	return &PageNavigateOutput{}, nil
}
