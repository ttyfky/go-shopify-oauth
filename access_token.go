package shopifyoauth

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/xerrors"
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
		ClientID:     c.App.ApiKey,
		ClientSecret: c.App.ApiSecret,
		Code:         code,
	}
}

func (c *Client) VerifyRedirect(r *http.Request, p *ShopAccessTokenParam) (bool, error) {
	// TODO: implement.
	urlVerified, err := c.App.VerifyAuthorizationURL(r.URL)
	if err != nil {
		return false, xerrors.Errorf("URL verification process failed %w", err)
	}
	if !urlVerified {
		return false, ErrURLVerification
	}
	if p.Nonce != c.ShopState[p.Shop] {
		return false, ErrStateVerification
	}
	return true, nil
}

func (c *Client) GetAccessToken(r *http.Request) (string, *ShopAccessTokenParam, error) {
	p := newShopAccessTokenParam(r.URL)

	verified, err := c.VerifyRedirect(r, p)
	if err != nil {
		return "", p, xerrors.Errorf("verification process failure: %w", err)
	}
	if !verified {
		return "", p, xerrors.Errorf("verification failure: %w", err)
	}

	token, err := c.App.GetAccessToken(p.Shop, p.Code)
	if err != nil {
		return "", p, err
	}
	return token, p, nil
}

func (c *Client) DefaultOAuthRedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if debugReq {
			_ = dumpRequest(os.Stdout, "token", r) // Ignore the error
		}
		log.Printf("Generating Access Token\n")

		token, p, err := c.GetAccessToken(r)
		if err != nil {
			log.Printf("Failed to get access token with param %+v due to %s", p.Shop, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully get access tokent for %s", p.Shop)

		err = c.TokenHandler.Handle(token)
		if err != nil {
			log.Printf("Failed to process token %s", err)
			w.WriteHeader(http.StatusInternalServerError)
		}

		destURL := fmt.Sprintf(DefaultDestinationURL, p.Shop)
		log.Printf("Redirecting to %s", destURL)

		w.Header().Set("Location", destURL)
		w.WriteHeader(http.StatusFound)
	}
}
