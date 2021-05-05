package shopifyoauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/xerrors"

	"github.com/google/go-querystring/query"
)

const AuthZURL = "https://%s/admin/oauth/authorize"

// AuthorizeOption is query params used on AuthZ redirect.
// [Shopify Docs](https://shopify.dev/tutorials/authenticate-with-oauth#step-2-ask-for-permission).
type AuthorizeOption struct {
	// Client ID is the API key.
	ClientID string `url:"client_id"`
	// Scope is comma separated value of AuthZ scope in Shopify.
	Scope string `url:"scope"`
	// RedirectURI is a url redirected after authorizing client.
	RedirectURI string `url:"redirect_uri"`
	// State is nonce to keep the AuthN secure.
	State string `url:"state"`
	// GrantOptions is [Access Mode](https://shopify.dev/concepts/about-apis/authentication#api-access-modes).
	GrantOptions string `url:"grant_options,omitempty"`
}

// newShopParam creates *ShopParam for authorize request.
func newShopParam(u *url.URL) *ShopParam {
	q := u.Query()
	return &ShopParam{
		HMAC:      q.Get(HMACParamKey),
		Shop:      q.Get(ShopParamKey),
		TimeStamp: q.Get(TimestampParamKey),
	}
}

func (c *Client) authorizeOption(sp *ShopParam) *AuthorizeOption {
	return &AuthorizeOption{
		ClientID:     c.App.APIKey,
		Scope:        c.App.Scope,
		RedirectURI:  fmt.Sprintf("%s%s", c.ServiceURL, RedirectPath),
		State:        c.ShopState[sp.Shop],
		GrantOptions: c.App.GrantOptions,
	}
}

// ShopParam is params sent from Shopify on AuthZ process start.
type ShopParam struct {
	HMAC      string
	Shop      string
	TimeStamp string
}

func (sp *ShopParam) verifyHMAC() bool {
	// TODO impliment
	return true
}

func (sp *ShopParam) verifyShop() bool {
	// TODO impliment
	return true
}

// PrepareRedirect generates an URL of Shopify for user's AuthZ view.
// While generating an URL, keep nonce per shop ID for CSRF prevention.
func (c *Client) PrepareRedirect(r *http.Request) (*url.URL, error) {
	sp := newShopParam(r.URL)

	c.CacheState(sp.Shop, randomString())
	v, err := query.Values(c.authorizeOption(sp))
	if err != nil {
		return nil, xerrors.Errorf("failed to generate params for redirect: %w", err)
	}
	authzURL, _ := url.Parse(fmt.Sprintf(AuthZURL, sp.Shop))
	authzURL.RawQuery = v.Encode()
	return authzURL, nil
}

// DefaultAuthorizeHandler is HTTP handler to respond AuthZ trigger to Shopify.
// A successful request is redirected to Shopify's App AuthZ view.
func (c *Client) DefaultAuthorizeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debugReq {
			_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
		}

		authzURL, err := c.PrepareRedirect(r)
		if err != nil {
			log.Printf("Failed to generate redirect URL for Shopify due to %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Location", authzURL.String())
		w.WriteHeader(http.StatusFound)
	}
}
