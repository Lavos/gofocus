package gofocus

import (
	"bytes"
	"github.com/nsf/termbox-go"
	"github.com/mrjones/oauth"
	"github.com/araddon/httpstream"
	"encoding/json"
	"log"
	"fmt"
	// "time"
)

type MicroTweet struct {
	Text, UserName, ScreenName, IDstr string
}

/* type LogEvent struct {
	Timestamp time.Time
	Message, Type string
} */

type Configuration struct {
	UserName, ConsumerKey, ConsumerSecret, Token, TokenSecret string
}

type Application struct {
	stream chan []byte
	key chan termbox.Event
	done chan bool
	value []rune

	position int
	reply_mode bool
	tweet_list []*MicroTweet

	log []s

	terminal *Terminal

	oc	*oauth.Consumer
	at	*oauth.AccessToken
}


func (a *Application) UpdateScreen() {
	var tweet MicroTweet
	var remaining int
	tp := &tweet

	if len(a.tweet_list) > 0 {
		tp = a.tweet_list[a.position]
		remaining = len(a.tweet_list) - 1 - a.position
	}

	a.terminal.DrawScreen(remaining, tp, a.value, a.reply_mode)
}

func (a *Application) InsertHandle() {
	if a.position <= len(a.tweet_list)-1 {
		current_tweet := a.tweet_list[a.position]
		a.value = []rune(fmt.Sprintf("@%v ", current_tweet.ScreenName))
	}
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
				}

			case termbox.KeyPgdn:
				if a.position < len(a.tweet_list)-1 {
					a.position++
					a.reply_mode = false
				}

			case termbox.KeyPgup:
				if a.position > 0 {
					a.position--
					a.reply_mode = false
				}

			case termbox.KeySpace:
				a.value = append(a.value, ' ')

			case termbox.KeyCtrlN:
				a.InsertHandle()

			case termbox.KeyCtrlR: // toggle reply mode
				if a.reply_mode {
					a.reply_mode = false
					a.value = a.value[len(a.value):]
				} else {
					a.InsertHandle()
					a.reply_mode = true
				}

			case termbox.KeyEnter:
				if len(a.value) > 0 {
					params := map[string]string{
						"status": string(a.value),
					}

					if a.reply_mode {
						params["in_reply_to_status_id"] = a.tweet_list[a.position].IDstr
					}

					a.oc.Post("https://api.twitter.com/1.1/statuses/update.json", params, a.at)
					a.value = a.value[len(a.value):]
					a.reply_mode = false
				}

			default:
				a.value = append(a.value, ev.Ch)
			}


		case b := <-a.stream:
			switch {
			case bytes.HasPrefix(b, []byte(`{"created_at":`)):

				tweet := httpstream.Tweet{}
				err := json.Unmarshal(b, &tweet)

				if err != nil {
					break
				}

				microTweet := MicroTweet{
					IDstr:		tweet.Id_str,
					Text:       tweet.Text,
					UserName:   tweet.User.Name,
					ScreenName: tweet.User.ScreenName,
				}

				a.tweet_list = append(a.tweet_list, &microTweet)
			}

		case <-a.done:
			termbox.Close()
			log.Print("Client lost connnection.")
			break loop
		}

		a.UpdateScreen()
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

	a.oc = oauth.NewConsumer(
		c.ConsumerKey,
		c.ConsumerSecret,
		oauth.ServiceProvider{
			RequestTokenUrl:   "https://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})

	httpstream.OauthCon = a.oc

	a.at = &oauth.AccessToken{
		Token:    c.Token,
		Secret:   c.TokenSecret,
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
