package shopifyoauth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
)

type Client struct {
	App          *App
	ServiceURL   string
	ShopState    map[string]string
	HTTPClient   *http.Client
	TokenHandler TokenHandler
}

type App struct {
	AppName      string
	APIKey       string
	APISecret    string
	Scope        string
	GrantOptions string
}

func NewClient(app *App, u string, th TokenHandle) *Client {
	return &Client{
		App:          app,
		ServiceURL:   u,
		ShopState:    make(map[string]string),
		HTTPClient:   http.DefaultClient,
		TokenHandler: th,
	}
}

func (c Client) CacheState(shop, state string) {
	c.ShopState[shop] = state
}

type TokenHandler interface {
	Handle(*AccessToken) error
}

type TokenHandle func(*AccessToken) error

func (th TokenHandle) Handle(token *AccessToken) error {
	return th(token)
}

// DefaultAccessTokenHandle is a default handler of AccessToken.
func DefaultAccessTokenHandle(at *AccessToken) error {
	log.Printf("got access token for scope %s", at.Scope)
	return nil
}

func isUnsuccessfulStatusCode(code int) bool {
	return http.StatusBadRequest <= code
}

func randomString() string {
	b := make([]byte, 64)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

const debugReq = false

func dumpRequest(writer io.Writer, header string, r *http.Request) error {
	data, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}
	_, _ = writer.Write([]byte("\n" + header + ": \n"))
	_, _ = writer.Write(data)
	return nil
}
