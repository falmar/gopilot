package gopilot

import (
	"context"
	"fmt"
	"time"

	"github.com/mafredri/cdp/protocol/domstorage"
	"github.com/mafredri/cdp/protocol/network"
	"github.com/mafredri/cdp/protocol/storage"
)

type PageStorage interface {
	// GetCookies retrieves cookies for the current page.
	// Takes a GetCookiesInput and returns GetCookiesOutput or an error.
	GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error)

	// SetCookies sets cookies for the current page.
	// Takes a SetCookiesInput and returns SetCookiesOutput or an error.
	SetCookies(ctx context.Context, in *SetCookiesInput) (*SetCookiesOutput, error)

	// ClearCookies clears cookies for the current page.
	ClearCookies(ctx context.Context) error

	GetLocalStorage(ctx context.Context, in *GetLocalStorageInput) (*GetLocalStorageOutput, error)
	SetLocalStorage(ctx context.Context, in *SetLocalStorageInput) (*SetLocalStorageOutput, error)
	ClearLocalStorage(ctx context.Context) error
}

// PageCookie represents a cookie in the browser.
// It includes details such as name, value, domain, path, expiration, and security features.
type PageCookie struct {
	Name     string     // The name of the cookie.
	Value    string     // The value of the cookie.
	Domain   string     // The domain the cookie is associated with.
	Path     string     // The path the cookie is accessible from.
	Size     int        // The size of the cookie in bytes.
	Expires  *time.Time // The expiration time of the cookie.
	Secure   bool       // Indicates if the cookie is secure (only sent over HTTPS).
	HttpOnly bool       // Indicates if the cookie is accessible via HTTP only (not accessible via JavaScript).
	Session  bool       // Indicates if the cookie is a session cookie.
}

// GetCookiesInput specifies the input for the GetCookies method.
type GetCookiesInput struct{}

// GetCookiesOutput contains the cookies retrieved from the browser.
// It returns a list of cookies.
type GetCookiesOutput struct {
	Cookies []PageCookie // List of cookies.
}

// GetCookies retrieves all cookies for the current page.
// Returns a GetCookiesOutput containing the cookies or an error if retrieval fails.
func (p *page) GetCookies(ctx context.Context, in *GetCookiesInput) (*GetCookiesOutput, error) {
	rp, err := p.client.Storage.GetCookies(ctx, &storage.GetCookiesArgs{})
	if err != nil {
		return nil, err
	}

	var cookies []PageCookie
	for _, c := range rp.Cookies {
		pc := PageCookie{
			Name:     c.Name,
			Value:    c.Value,
			Domain:   c.Domain, // Corrected to use Domain from cookie
			Path:     c.Path,   // Corrected to use Path from cookie
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

// SetCookiesInput specifies the input for the SetCookies method.
// It contains a list of cookies to set in the browser.
type SetCookiesInput struct {
	Cookies []PageCookie // List of cookies to set.
}

// SetCookiesOutput is returned after setting cookies successfully.
type SetCookiesOutput struct{}

// SetCookies sets the specified cookies in the browser.
// Returns a SetCookiesOutput or an error if setting fails.
func (p *page) SetCookies(ctx context.Context, in *SetCookiesInput) (*SetCookiesOutput, error) {
	var cookies []network.CookieParam
	for _, c := range in.Cookies {
		ncp := network.CookieParam{
			Name:  c.Name,
			Value: c.Value,
		}
		if c.Domain != "" {
			ncp.Domain = &c.Domain
		}
		if c.Path != "" {
			ncp.Path = &c.Path
		}
		if c.Secure {
			ncp.Secure = &c.Secure
		}
		if c.HttpOnly {
			ncp.HTTPOnly = &c.HttpOnly
		}
		if c.Expires != nil {
			ncp.Expires = network.TimeSinceEpoch(c.Expires.Unix())
		}
		cookies = append(cookies, ncp)
	}

	err := p.client.Storage.SetCookies(ctx, &storage.SetCookiesArgs{Cookies: cookies})
	if err != nil {
		return nil, err
	}
	return &SetCookiesOutput{}, nil
}

// ClearCookies clears all cookies from the browser.
func (p *page) ClearCookies(ctx context.Context) error {
	return p.client.Storage.ClearCookies(ctx, &storage.ClearCookiesArgs{})
}

type LocalStorageItem struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type SetLocalStorageInput struct {
	Items []LocalStorageItem
}
type SetLocalStorageOutput struct{}

func (p *page) SetLocalStorage(ctx context.Context, in *SetLocalStorageInput) (*SetLocalStorageOutput, error) {
	pageURL, err := getPageCurrentURL(ctx, p)
	if err != nil {
		return nil, err
	}
	origin := fmt.Sprintf("%s://%s", pageURL.Scheme, pageURL.Host)

	for _, i := range in.Items {
		err := p.client.DOMStorage.SetDOMStorageItem(ctx, &domstorage.SetDOMStorageItemArgs{
			StorageID: domstorage.StorageID{
				IsLocalStorage: true,
				SecurityOrigin: &origin,
			},
			Key:   i.Name,
			Value: i.Value,
		})
		if err != nil {
			return nil, err
		}
	}

	return &SetLocalStorageOutput{}, nil
}

type GetLocalStorageInput struct{}
type GetLocalStorageOutput struct {
	Items []LocalStorageItem
}

func (p *page) GetLocalStorage(ctx context.Context, in *GetLocalStorageInput) (*GetLocalStorageOutput, error) {
	pageURL, err := getPageCurrentURL(ctx, p)
	if err != nil {
		return nil, err
	}
	origin := fmt.Sprintf("%s://%s", pageURL.Scheme, pageURL.Host)

	rp, err := p.client.DOMStorage.GetDOMStorageItems(ctx, &domstorage.GetDOMStorageItemsArgs{
		StorageID: domstorage.StorageID{
			IsLocalStorage: true,
			SecurityOrigin: &origin,
		},
	})
	if err != nil {
		return nil, err
	}

	var items []LocalStorageItem
	for _, i := range rp.Entries {
		items = append(items, LocalStorageItem{
			Name:  i[0],
			Value: i[1],
		})
	}

	return &GetLocalStorageOutput{Items: items}, nil
}

type ClearLocalStorageInput struct{}
type ClearLocalStorageOutput struct{}

func (p *page) ClearLocalStorage(ctx context.Context) error {
	pageURL, err := getPageCurrentURL(ctx, p)
	if err != nil {
		return err
	}
	origin := fmt.Sprintf("%s://%s", pageURL.Scheme, pageURL.Host)

	return p.client.DOMStorage.Clear(ctx, &domstorage.ClearArgs{
		StorageID: domstorage.StorageID{IsLocalStorage: true, SecurityOrigin: &origin},
	})
}
