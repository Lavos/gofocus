package main

import (
	"log"
	"fmt"
	oauth "github.com/araddon/goauth"

	"github.com/kelseyhightower/envconfig"
)

var (

)

const (

)

type Config struct {
	ConsumerKey, ConsumerSecret string
}

func main () {
	c := &Config{}
	envconfig.Process("gofocus", c)

	oc := &oauth.OAuthConsumer{
		Service:          "twitter",
		RequestTokenURL:  "https://api.twitter.com/oauth/request_token",
		AccessTokenURL:   "https://api.twitter.com/oauth/access_token",
		AuthorizationURL: "https://api.twitter.com/oauth/authorize",
		ConsumerKey:      c.ConsumerKey,
		ConsumerSecret:   c.ConsumerSecret,
		CallBackURL:      "oob",
		UserAgent:        "go/httpstream",
	}

	url, req_token, err := oc.GetRequestAuthorizationURL()

	log.Printf("Authorization URL: %#v", url)
	log.Printf("Requestion Token: %#v", req_token)
	log.Printf("Error: %#v", err)

	var pin string
	fmt.Print("Validator PIN: ")
	fmt.Scanln(&pin)

	access_token := oc.GetAccessToken(req_token.Token, pin)

	log.Printf("access_token: %#v", access_token)
}
