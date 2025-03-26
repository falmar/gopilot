package gopilot

import (
	"context"

	"github.com/mafredri/cdp/protocol/dom"
)

// GetContent retrieves the HTML content of the current page.
// It returns the outer HTML as a string or an error if retrieval fails.
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

// PageQuerySelectorInput contains the selector string for querying elements.
type PageQuerySelectorInput struct {
	Selector string
}

// PageQuerySelectorOutput contains the Element found by the query.
type PageQuerySelectorOutput struct {
	Element Element
}

// QuerySelector finds an element in the page that matches the given CSS selector.
// It returns a PageQuerySelectorOutput containing the Element or an error if the query fails.
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
		Element: newElement(drp.Node, p.client),
	}, nil
}
