package gopilot

import (
	"context"
	"errors"

	"github.com/mafredri/cdp/protocol/dom"
)

var ErrElementNotFound = errors.New("element not found")

type PageDOM interface {
	// GetContent retrieves the HTML content of the page as a string.
	// Returns the content or an error if retrieving fails.
	GetContent(ctx context.Context) (string, error)

	// SetContent replaces the current DOM with supplied content
	SetContent(ctx context.Context, content string) error

	// QuerySelector finds an element matching the selector.
	// Takes a PageQuerySelectorInput and returns a PageQuerySelectorOutput or an error.
	QuerySelector(ctx context.Context, in *PageQuerySelectorInput) (*PageQuerySelectorOutput, error)

	// Search finds an element matching the text, query selector or xpath
	// Takes a PageSearchInput and returns a PageSearchOutput or an error.
	Search(ctx context.Context, in *PageSearchInput) (*PageSearchOutput, error)
}

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

// SetContent replaces the current DOM with supplied content
func (p *page) SetContent(ctx context.Context, content string) error {
	doc, err := p.client.DOM.GetDocument(ctx, nil)
	if err != nil {
		return err
	}

	return p.client.DOM.SetOuterHTML(ctx, &dom.SetOuterHTMLArgs{
		NodeID:    doc.Root.NodeID,
		OuterHTML: content,
	})
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

	if qrp.NodeID == 0 {
		return nil, ErrElementNotFound
	}

	drp, err := p.client.DOM.DescribeNode(ctx, &dom.DescribeNodeArgs{
		NodeID: &qrp.NodeID,
	})
	if err != nil {
		return nil, err
	}

	rrp, err := p.client.DOM.ResolveNode(ctx, &dom.ResolveNodeArgs{
		NodeID: &qrp.NodeID,
	})
	if err != nil {
		return nil, err
	}

	return &PageQuerySelectorOutput{
		Element: newElement(drp.Node, rrp.Object, p.client),
	}, nil
}

// PageSearchInput contains the selector string for querying elements.
type PageSearchInput struct {
	Selector string
	Pierce   bool
}

// PageSearchOutput contains the Element found by the query.
type PageSearchOutput struct {
	Element Element
}

func (p *page) Search(ctx context.Context, in *PageSearchInput) (*PageSearchOutput, error) {
	_, err := p.client.DOM.GetDocument(ctx, nil)
	if err != nil {
		return nil, err
	}

	qsrp, err := p.client.DOM.PerformSearch(ctx, &dom.PerformSearchArgs{
		Query:                     in.Selector,
		IncludeUserAgentShadowDOM: &in.Pierce,
	})
	if err != nil {
		return nil, err
	} else if qsrp.ResultCount <= 0 {
		return nil, ErrElementNotFound
	}

	srp, err := p.client.DOM.GetSearchResults(ctx, &dom.GetSearchResultsArgs{
		SearchID:  qsrp.SearchID,
		FromIndex: 0,
		ToIndex:   qsrp.ResultCount,
	})
	if err != nil {
		return nil, err
	}

	var firstMatch dom.NodeID = 0

	for _, id := range srp.NodeIDs {
		if id != 0 {
			firstMatch = id
			break
		}
	}

	if firstMatch == 0 {
		return nil, ErrElementNotFound
	}

	drp, err := p.client.DOM.DescribeNode(ctx, &dom.DescribeNodeArgs{
		NodeID: &firstMatch,
	})
	if err != nil {
		return nil, err
	}

	rrp, err := p.client.DOM.ResolveNode(ctx, &dom.ResolveNodeArgs{
		NodeID: &drp.Node.NodeID,
	})
	if err != nil {
		return nil, err
	}

	return &PageSearchOutput{
		Element: newElement(drp.Node, rrp.Object, p.client),
	}, nil
}
