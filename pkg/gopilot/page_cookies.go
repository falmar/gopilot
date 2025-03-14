package gopilot

import (
	"context"
	"errors"
	"time"
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
	Cookies PageCookie
}

func (p *page) GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error) {
	return nil, errors.New("not implemented")
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
