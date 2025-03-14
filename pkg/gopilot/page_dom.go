package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/dom"
)

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

type PageQuerySelectorInput struct {
	Selector string
}
type PageQuerySelectorOutput struct {
	Element Element
}

func (p *page) QuerySelector(ctx context.Context, in *PageQuerySelectorInput) (*PageQuerySelectorOutput, error) {
	doc, err := p.client.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	qrp, err := p.client.DOM.QuerySelector(ctx, &dom.QuerySelectorArgs{
		NodeID:   doc.Root.NodeID,
		Selector: in.Selector,
	})
	if err != nil {
		return nil, err
	}

	drp, err := p.client.DOM.DescribeNode(ctx, &dom.DescribeNodeArgs{
		NodeID: &qrp.NodeID,
	})
	if err != nil {
		return nil, err
	}

	return &PageQuerySelectorOutput{
		Element: newElement(drp.Node, p.devtool, p.client),
	}, nil
}
