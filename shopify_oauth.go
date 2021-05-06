package shopifyoauth

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/http/httputil"

	goshopify "github.com/bold-commerce/go-shopify/v3"
)

type Client struct {
	App          *goshopify.App
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

func NewClient(app *goshopify.App, th TokenHandle) *Client {
	return &Client{
		App:          app,
		ShopState:    make(map[string]string),
		HTTPClient:   http.DefaultClient,
		TokenHandler: th,
	}
}

func (c Client) CacheState(shop, state string) {
	c.ShopState[shop] = state
}

type TokenHandler interface {
	Handle(string) error
}

type TokenHandle func(string) error

func (th TokenHandle) Handle(token string) error {
	return th(token)
}

// DefaultAccessTokenHandle is a default handler of AccessToken.
func DefaultAccessTokenHandle(token string) error {
	log.Printf("Got access token: len(token) = %d", len(token))
	return nil
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
