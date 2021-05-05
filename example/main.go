package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	shopifyoauth "github.com/ttyfky/go-shopify-oauth"
)

func main() {
	h := shopifyoauth.NewClient(
		&shopifyoauth.App{
			AppName:      os.Getenv("APP_NAME"),
			APIKey:       os.Getenv("API_KEY"),
			APISecret:    os.Getenv("API_SECRET"),
			Scope:        os.Getenv("SCOPE"),
			GrantOptions: os.Getenv("GRANT_OPTION"),
		},
		os.Getenv("SERVICE_URL"),
		shopifyoauth.DefaultAccessTokenHandle,
	)
	http.HandleFunc("/", h.DefaultAuthorizeHandler())
	http.HandleFunc(shopifyoauth.RedirectPath, h.DefaultOAuthRedirectHandler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", 8090), nil))
}
