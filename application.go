package gofocus

import (
	"bytes"
	"github.com/nsf/termbox-go"
	oauth "github.com/araddon/goauth"
	"github.com/araddon/httpstream"
	"encoding/json"
	"log"
)

type MicroTweet struct {
	Text, UserName, ScreenName string
}

type Configuration struct {
	UserName, ConsumerKey, ConsumerSecret, Token, TokenSecret string
}

type Application struct {
	stream chan []byte
	key chan termbox.Event
	done chan bool
	value []rune

	position int
	tweet_list []*MicroTweet

	terminal *Terminal

	oc	*oauth.OAuthConsumer
	at	*oauth.AccessToken
}


func (a *Application) Run(){
loop:
	for {
		select {
		case ev := <-a.key:
			switch ev.Key {
			case termbox.KeyEsc:
				termbox.Close()
				break loop

			case termbox.KeyBackspace, termbox.KeyBackspace2:
				if len(a.value) > 0 {
					a.value = a.value[:len(a.value)-1]
					a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)
				}

			case termbox.KeyPgdn:
				if a.position < len(a.tweet_list)-1 {
					a.position++
					a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)
				}

			case termbox.KeyPgup:
				if a.position > 0 {
					a.position--
					a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)
				}

			case termbox.KeySpace:
				a.value = append(a.value, ' ')
				a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)

			case termbox.KeyEnter:
				if len(a.value) > 0 {
					r, err := a.oc.Post("https://api.twitter.com/1.1/statuses/update.json",
						oauth.Params{
							&oauth.Pair{Key:"status", Value: string(a.value)},
						}, a.at)

					log.Printf("response: %#v", r)
					log.Printf("err: %#v", err)
					log.Printf("value: %#v", string(a.value))

					a.value = a.value[len(a.value):]
				}

			default:
				a.value = append(a.value, ev.Ch)
				a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)
			}

		case b := <-a.stream:
			switch {
			case bytes.HasPrefix(b, []byte(`{"created_at":`)):
				// log.Print("%#v", b)

				tweet := httpstream.Tweet{}
				err := json.Unmarshal(b, &tweet)

				if err != nil {
					break
				}

				microTweet := MicroTweet{
					Text:       tweet.Text,
					UserName:   tweet.User.Name,
					ScreenName: tweet.User.ScreenName,
				}

				a.tweet_list = append(a.tweet_list, &microTweet)
				a.terminal.DrawScreen(len(a.tweet_list) - a.position, a.tweet_list[a.position], a.value)
			}

		case <-a.done:
			termbox.Close()
			log.Print("Client lost connnection.")
			break loop
		}
	}
}

func NewApplication(c *Configuration) *Application {
	a := &Application{
		stream: make(chan []byte, 1000),
		key: make(chan termbox.Event),
		done: make(chan bool),
		value: make([]rune, 0),
		terminal: NewTerminal(),
	}

	a.oc = &oauth.OAuthConsumer{
		Service:          "twitter",
		RequestTokenURL:  "http://twitter.com/oauth/request_token",
		AccessTokenURL:   "http://twitter.com/oauth/access_token",
		AuthorizationURL: "http://twitter.com/oauth/authorize",
		ConsumerKey:      c.ConsumerKey,
		ConsumerSecret:   c.ConsumerSecret,
		CallBackURL:      "oob",
		UserAgent:        "go/httpstream",
	}

	httpstream.OauthCon = a.oc

	a.at = &oauth.AccessToken{
		Id:       "",
		Token:    c.Token,
		Secret:   c.TokenSecret,
		UserRef:  c.UserName,
		Verifier: "",
		Service:  "twitter",
	}

	client := httpstream.NewOAuthClient(a.at, httpstream.OnlyTweetsFilter(func(line []byte) {
		a.stream <- line
	}))

	client.SetMaxWait(5)

	// client.Sample(a.done)
	client.User(a.done)

	go a.terminal.Run(a.key)

	return a
}