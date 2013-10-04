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
)

type MicroTweet struct {
	Text string
	UserName string
	ScreenName string
}

var (
	position = 0
	tweet_list = make([]*MicroTweet, 0)
)

const (
	ConsumerKey = "wLHVnPDOveYQnshdIUt97w"
	ConsumerSecret = "RZkPGW49NaWCZpkBw8MxA3SP2EOOxwa9HVRgre7Yo"

	Token = "7938332-PN1C2Gxqycd5igYFotzHdlmzsZQ0cKmQyiZ2iII"
	TokenSecret = "DD4oidwbwvfXTl6ddiBeI6OrURFFipWijeblsRFtZ0"
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
	print_tb(1, 0, termbox.ColorBlack | termbox.AttrBold, termbox.ColorWhite, "Go Focus")

	if (position >= len(tweet_list)) {
		return
	}

	tweet := tweet_list[position]

	// status line
	color_line(1, termbox.ColorYellow)
	print_tb(1, 1, termbox.ColorBlack, termbox.ColorYellow, fmt.Sprintf("%v tweets remaining.", len(tweet_list) - position -1))

	// tweet
	print_tb(1, 2, termbox.ColorMagenta, termbox.ColorBlack, tweet.ScreenName)
	print_tb(21, 2, termbox.ColorGreen, termbox.ColorBlack, tweet.UserName)
	print_tb(1, 3, termbox.ColorWhite, termbox.ColorBlack, tweet.Text)
}

func main () {
	stream := make(chan []byte, 1000)
	key := make(chan termbox.Key)
	done := make(chan bool)

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
		ConsumerKey:      ConsumerKey,
		ConsumerSecret:   ConsumerSecret,
		CallBackURL:      "oob",
		UserAgent:        "go/httpstream",
	}

	at := oauth.AccessToken{
		Id: "",
		Token:    Token,
		Secret:   TokenSecret,
		UserRef:  "lavos",
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
			// log.Printf("%#v", string(b))

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
