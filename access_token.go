package shopifyoauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/xerrors"

	"github.com/google/go-querystring/query"
)

// AccessTokenURL is the URL to extract access token of Shopify.
const AccessTokenURL = "https://%s/admin/oauth/access_token"

// DefaultDestinationURL is an URL of redirect target after App's installation.
const DefaultDestinationURL = "https://%s/admin/apps"

// AccessTokenRequestPayload is the payload of oauth access_token endpoint.
type AccessTokenRequestPayload struct {
	// Client ID is the API key.
	ClientID string `url:"client_id"`
	// ClientSecret is the API secret.
	ClientSecret string `url:"client_secret"`
	// code is the authorization code provided in the redirect.
	Code string `url:"code"`
}

// AccessToken is the response of oauth access_token endpoint.
type AccessToken struct {
	Code  string `json:"code"`
	Scope string `json:"scope"`
}

// OnlineModeAccessToken is the response of oauth access_token endpoint when AuthZ was for online mode.
type OnlineModeAccessToken struct {
	*AccessToken
	ExpiresIn string `json:"expires_in"`

	AssociatedUser struct {
		ID            string `json:"id"`
		FirstName     string `json:"first_name"`
		LastName      string `json:"last_name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		AccountOwner  bool   `json:"account_owner"`
		Locale        string `json:"locale"`
		Collaborator  bool   `json:"collaborator"`
	} `json:"associated_user"`
}

// ShopAccessTokenParam is used to store params used in redirect request from Shopify.
type ShopAccessTokenParam struct {
	*ShopParam
	Nonce string
	Host  string
	Code  string
}

func newShopAccessTokenParam(u *url.URL) *ShopAccessTokenParam {
	q := u.Query()
	return &ShopAccessTokenParam{
		ShopParam: &ShopParam{
			HMAC:      q.Get(HMACParamKey),
			Shop:      q.Get(ShopParamKey),
			TimeStamp: q.Get(TimestampParamKey),
		},
		Nonce: q.Get(StateParamKey),
		Host:  q.Get(HostParamKey),
		Code:  q.Get(CodeParamKey),
	}
}

func (c *Client) accessTokenRequestPayload(code string) *AccessTokenRequestPayload {
	return &AccessTokenRequestPayload{
		ClientID:     c.App.APIKey,
		ClientSecret: c.App.APISecret,
		Code:         code,
	}
}

func (s *ShopAccessTokenParam) Validate(nonce string) bool {
	// TODO: implement.
	return s.verifyHMAC() && s.verifyShop() && nonce == s.Nonce
}

func (c *Client) GetAccessToken(r *http.Request) (*AccessToken, *ShopAccessTokenParam, error) {
	p := newShopAccessTokenParam(r.URL)
	if !p.Validate(c.ShopState[p.Shop]) {
		return nil, p, xerrors.Errorf("request from Shopify had invalid params")
	}

	at, err := c.getAccessTokenFromShopify(p)
	if err != nil {
		return nil, p, err
	}
	return at, p, nil
}

func (c *Client) getAccessTokenFromShopify(p *ShopAccessTokenParam) (*AccessToken, error) {
	v, err := query.Values(c.accessTokenRequestPayload(p.Code))
	if err != nil {
		return nil, xerrors.Errorf("failed to decode auth token payload: %w", err)
	}

	res, err := c.HTTPClient.PostForm(fmt.Sprintf(AccessTokenURL, p.Shop), v)
	if err != nil {
		return nil, xerrors.Errorf("failed to request to access_token: %w", err)
	}

	if isUnsuccessfulStatusCode(res.StatusCode) {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, xerrors.Errorf("got unsuccessful response code %d but failed to read response body: %w", res.StatusCode, err)
		}
		return nil, xerrors.Errorf("unsuccessful request to access_token endpoint: status: %d, msg: %s", res.StatusCode, string(b))
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, xerrors.Errorf("failed to read response body; %w", err)
	}

	at := &AccessToken{}
	err = json.Unmarshal(b, at)
	if err != nil {
		return nil, xerrors.Errorf("failed to unmarshal AccessToken: %w", err)
	}
	return at, nil
}

func (c *Client) DefaultOAuthRedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debugReq {
			_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
		}

		at, p, err := c.GetAccessToken(r)
		if err != nil {
			log.Printf("Failed to get access token with param %+v due to %s", p.Shop, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// TODO: add process access token
		log.Printf("%+v\n", at)
		w.Header().Set("Location", fmt.Sprintf(DefaultDestinationURL, p.Shop))
		w.WriteHeader(http.StatusFound)
	}
}
