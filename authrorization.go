package shopifyoauth

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/xerrors"
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

// ShopParam is params sent from Shopify on AuthZ process start.
type ShopParam struct {
	HMAC      string `url:"hmac"`
	Shop      string `url:"shop"`
	TimeStamp string `url:"timestamp"`
}

// PrepareRedirect generates an URL of Shopify for user's AuthZ view.
// While generating an URL, keep nonce per shop ID for CSRF prevention.
func (c *Client) PrepareRedirect(r *http.Request) (string, error) {
	verified, err := c.App.VerifyAuthorizationURL(r.URL)
	if err != nil {
		return "", err
	}
	if !verified {
		return "", ErrURLVerification
	}
	sp := newShopParam(r.URL)
	state := randomString()
	c.CacheState(sp.Shop, state)
	return c.App.AuthorizeUrl(sp.Shop, state), nil
}

// DefaultAuthorizeHandler is HTTP handler to respond AuthZ trigger to Shopify.
// A successful request is redirected to Shopify's App AuthZ view.
func (c *Client) DefaultAuthorizeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debugReq {
			_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
		}
		log.Printf("Processing request of %s", r.URL.RawPath)
		authzURL, err := c.PrepareRedirect(r)
		if err != nil {
			if xerrors.Is(err, ErrURLVerification) {
				log.Printf("URL unverified %s", r.URL)
				w.WriteHeader(http.StatusBadRequest)
			}
			log.Printf("Failed to generate redirect URL for Shopify due to %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Redirecting to %s", authzURL)
		w.Header().Set("Location", authzURL)
		w.WriteHeader(http.StatusFound)
	}
}
