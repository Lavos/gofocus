package main

import (
	"github.com/araddon/httpstream"
	oauth "github.com/araddon/goauth"
	"github.com/nsf/termbox-go"
	"log"
	"fmt"
	"encoding/json"
	"strings"
	"bytes"
	"flag"
	"os"
)

type MicroTweet struct {
	Text, UserName, ScreenName string
}

type Configuration struct {
	UserName, ConsumerKey, ConsumerSecret, Token, TokenSecret string
}

var (
	position = 0
	tweet_list = make([]*MicroTweet, 0)
	config_filename = flag.String("c", "", "filename of json configuration file")
)

func color_line (y int, bg termbox.Attribute) {
	var width, _ = termbox.Size()

	for c := 0; c < width; c++ {
		termbox.SetCell(c, y, '\000', termbox.ColorDefault, bg)
	}
}

func print_tb (x, y int, fg, bg termbox.Attribute, msg string) {
	var width, _ = termbox.Size()
	var clean string

	clean = strings.Replace(msg, "\n", "", -1)

	for _, c := range clean {
		if x >= width {
			x = 0
			y++
		}

		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func drawScreen () {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	// title line
	color_line(0, termbox.ColorWhite)
	print_tb(1, 0, termbox.ColorBlack, termbox.ColorWhite, "Go Focus")

	remaining := len(tweet_list) - position
	var bar_color termbox.Attribute

	if remaining == 0 {
		bar_color = termbox.ColorWhite
	} else {
		bar_color = termbox.ColorYellow
	}

	// status line
	color_line(1, bar_color)
	print_tb(1, 1, termbox.ColorBlack, bar_color, fmt.Sprintf("%v tweets remaining.", remaining))

	if len(tweet_list) == 0 {
		return
	}

	// tweet
	tweet := tweet_list[position]
	print_tb(1, 2, termbox.ColorMagenta, termbox.ColorBlack, tweet.ScreenName)
	print_tb(21, 2, termbox.ColorGreen, termbox.ColorBlack, tweet.UserName)
	print_tb(1, 3, termbox.ColorWhite, termbox.ColorBlack, tweet.Text)
}

func main () {
	flag.Parse()

	stream := make(chan []byte, 1000)
	key := make(chan termbox.Key)
	done := make(chan bool)

	log.Printf("[config_filename] %#v", *config_filename)

	// configuration JSON
	config_file, err := os.Open(*config_filename)

	if err != nil {
		log.Fatal(err)
	}

	data := make([]byte, 2048)
	n, err := config_file.Read(data)

	if err != nil {
		log.Fatal(err)
	}

	var config Configuration
	err = json.Unmarshal(data[:n], &config)

	if err != nil {
		log.Fatal(err)
	}

	// termbox
	tb_err := termbox.Init()
	if tb_err != nil {
		log.Panic(tb_err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)
	httpstream.OauthCon = &oauth.OAuthConsumer{
		Service:          "twitter",
		RequestTokenURL:  "http://twitter.com/oauth/request_token",
		AccessTokenURL:   "http://twitter.com/oauth/access_token",
		AuthorizationURL: "http://twitter.com/oauth/authorize",
		ConsumerKey:      config.ConsumerKey,
		ConsumerSecret:   config.ConsumerSecret,
		CallBackURL:      "oob",
		UserAgent:        "go/httpstream",
	}

	at := oauth.AccessToken{
		Id: "",
		Token:    config.Token,
		Secret:   config.TokenSecret,
		UserRef:  config.UserName,
		Verifier: "",
		Service:  "twitter",
	}

	client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
		stream <- line
	}))

	client.User(done)

	go func () {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				key <- ev.Key
			}
		}
	}()

	drawScreen()
loop:
	for {
		select {
		case k := <-key:
			switch k {
			case termbox.KeyEsc:
				break loop

			case termbox.KeySpace:
				if position < len(tweet_list) -1 {
					position++
					drawScreen()
				}
			}

		case b := <-stream:
			switch {
			case bytes.HasPrefix(b, []byte(`{"created_at":`)):
				tweet := httpstream.Tweet{}
				err := json.Unmarshal(b, &tweet)

				if err != nil {
					break
				}

				microTweet := MicroTweet{
					Text: tweet.Text,
					UserName: tweet.User.Name,
					ScreenName: tweet.User.ScreenName,
				}

				tweet_list = append(tweet_list, &microTweet)
				drawScreen()
			}
		}
	}
}
