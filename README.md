# go-shopify-oauth
OAuth Server for Shopify App.

# How OAuth of Shopify App works
How to authenticate with OAuth for Shopify App is described in [here](https://shopify.dev/tutorials/authenticate-with-oaut).

In Short, OAuth client needs to provide two endpoints:

1. Endpoint to provide App's detail 
   * Provide requesting scope of App for authorization with a shop
    
1. Endpoint to handle redirect after authorization
   * Receive AuthZ code so that requesting access token
    
# Example
Runnable server example is in [example](/example/main.go). 

## Prerequisite
Following env var is required to run the example.

* APP_NAME
  * Application's name
* API_KEY 
  * API key of Shopify App
* API_SECRET
  * API Secret of Shopify App
* SCOPE
  * OAuth Scope of Shopify App
* GRANT_OPTION
  * (Optional) [Access mode](https://shopify.dev/concepts/about-apis/authentication#api-access-modes). Default is Offline
* SERVICE_URL
  * URL of the service 


## Installation Steps of Shopify App
1. Create App in Shopify Partner dashboard
1. Get `API key` and `API Secret key` from App Setup page and set them in env var
1. Set env var of Scopes that an App needs
1. Set env var of host name where your service will run 
1. Run the service
1. Go to `Test on development store` in App page of  Shopify Partner dashboard
1. Select your test store and proceed installation

