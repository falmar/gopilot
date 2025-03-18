package gopilot

import (
	"context"
	"errors"
	"time"

	"github.com/mafredri/cdp/protocol/storage"
)

type PageCookie struct {
	Name   string
	Value  string
	Domain string
	Path   string

	Size    int
	Expires *time.Time

	Secure   bool
	HttpOnly bool
	Session  bool
}

type GetCookiesInput struct{}
type GetCookiesOutput struct {
	Cookies []*PageCookie
}

func (p *page) GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error) {
	rp, err := p.client.Storage.GetCookies(ctx, &storage.GetCookiesArgs{})
	if err != nil {
		return nil, err
	}

	var cookies []*PageCookie

	for _, c := range rp.Cookies {
		pc := &PageCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Name,
			Path:     c.Name,
			Size:     c.Size,
			Expires:  nil,
			Secure:   c.Secure,
			HttpOnly: c.HTTPOnly,
			Session:  c.Session,
		}

		if c.Expires > 0 {
			t := time.Unix(int64(c.Expires), 0)
			pc.Expires = &t
		}

		cookies = append(cookies, pc)
	}

	return &GetCookiesOutput{Cookies: cookies}, nil
}

type SetCookiesInput struct {
	Cookies PageCookie
}
type SetCookiesOutput struct{}

func (p *page) SetCookies(ctx context.Context, in *SetCookiesInput) (*SetCookiesOutput, error) {
	return nil, errors.New("not implemented")
}

type ClearCookiesInput struct {
	Cookies PageCookie
}
type ClearCookiesOutput struct{}

func (p *page) ClearCookies(ctx context.Context, in *ClearCookiesInput) (*ClearCookiesOutput, error) {
	return nil, errors.New("not implemented")
}
