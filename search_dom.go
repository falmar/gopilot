package gopilot

import (
	"context"
	"time"

	"github.com/mafredri/cdp"
	"github.com/mafredri/cdp/protocol/dom"
)

type searchInput struct {
	selector          string
	nodeID            int
	pierce            bool
	breakOnFirstMatch bool
	wait              time.Duration
	tick              time.Duration
}

func search(ctx context.Context, client *cdp.Client, in searchInput) ([]dom.Node, error) {
	var qsrp *dom.PerformSearchReply

	// Normalize wait and tick values
	wait := in.wait
	tick := in.tick
	if wait <= 0 {
		wait = 0
		tick = 0
	} else if tick <= 0 {
		tick = time.Second
	}

	tm := time.NewTimer(wait)
	defer tm.Stop()

	firstTry := true

waitLoop:
	for {
		var tk *time.Timer
		if firstTry {
			firstTry = false
			tk = time.NewTimer(0)
		} else {
			tk = time.NewTimer(tick)
		}

		select {
		case <-ctx.Done():
			tk.Stop()
			return nil, ctx.Err()
		case <-tm.C:
			tk.Stop()
			if wait > 0 {
				return nil, ErrElementSearchTimeout
			}
			return nil, ErrElementNotFound
		case <-tk.C:
			tk.Stop()
			_, err := client.DOM.GetDocument(ctx, nil)
			if err != nil {
				return nil, err
			}

			qsrp, err = client.DOM.PerformSearch(ctx, &dom.PerformSearchArgs{
				Query:                     in.selector,
				IncludeUserAgentShadowDOM: &in.pierce,
			})
			if err != nil {
				return nil, err
			} else if qsrp.ResultCount <= 0 {
				if wait <= 0 {
					return nil, ErrElementNotFound
				}
				continue
			}

			break waitLoop
		}
	}

	srp, err := client.DOM.GetSearchResults(ctx, &dom.GetSearchResultsArgs{
		SearchID:  qsrp.SearchID,
		FromIndex: 0,
		ToIndex:   qsrp.ResultCount,
	})
	if err != nil {
		return nil, err
	}

	result := make([]dom.Node, 0, len(srp.NodeIDs))
	for _, id := range srp.NodeIDs {
		if id == 0 {
			continue
		}
		drp, err := client.DOM.DescribeNode(
			ctx,
			&dom.DescribeNodeArgs{
				NodeID: &id,
			},
		)
		if err != nil {
			continue
		}
		result = append(result, drp.Node)
		if in.breakOnFirstMatch {
			break
		}
	}

	return result, nil
}
