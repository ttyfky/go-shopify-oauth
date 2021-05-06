package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	goshopify "github.com/bold-commerce/go-shopify/v3"

	shopifyoauth "github.com/ttyfky/go-shopify-oauth"
)

func main() {
	h := shopifyoauth.NewClient(
		&goshopify.App{
			ApiKey:      os.Getenv("API_KEY"),
			ApiSecret:   os.Getenv("API_SECRET"),
			Scope:       os.Getenv("SCOPE"),
			RedirectUrl: os.Getenv("SERVICE_URL") + shopifyoauth.RedirectPath,
		},
		shopifyoauth.DefaultAccessTokenHandle,
	)
	http.HandleFunc("/", h.DefaultAuthorizeHandler())
	http.HandleFunc(shopifyoauth.RedirectPath, h.DefaultOAuthRedirectHandler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8090), nil))
}
