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

func color_hline (y int, bg termbox.Attribute) {
	var width, _ = termbox.Size()

	for c := 0; c < width; c++ {
		termbox.SetCell(c, y, '\000', termbox.ColorDefault, bg)
	}
}

func print_wordwrap (msg string, width, x int) {
	working := make([]byte, 0)
	lines := make([][]byte, 0)

	working = append(working, msg...)

work:
	for len(working) != 0 {

		// if the width is less than a line, no problem!
		if len(working) <= width {
			lines = append(lines, working)
			working = working[len(working):]

			continue;
		}

		// find the space closest to the end, and split there
		for xw := width; xw != 0; xw-- {
			if working[xw] == 0x20 {
				lines = append(lines, working[:xw])
				working = working[xw+1:]
				continue work;
			}
		}

		// didn't find any spaces to split on, so just split at width
		lines = append(lines, working[:width])
		working = working[width:]
	}

	for index, line := range lines {
		print_line(1, index+x, termbox.ColorWhite, termbox.ColorDefault, string(line))
	}
}

func print_line (x, y int, fg, bg termbox.Attribute, msg string) {
	var clean string = strings.Replace(msg, "\n", "", -1)

	for _, c := range clean {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}

	// termbox.Flush()
}

func drawScreen () {
	var remaining int = 0
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	defer termbox.Flush()

	// title line
	color_hline(0, termbox.ColorWhite)
	print_line(1, 0, termbox.ColorBlack, termbox.ColorWhite, "Go Focus")

	if len(tweet_list) != 0 {
		remaining = len(tweet_list) - position - 1
	}

	var bar_color termbox.Attribute

	if remaining == 0 {
		bar_color = termbox.ColorWhite
	} else {
		bar_color = termbox.ColorYellow
	}

	// status line
	color_hline(1, bar_color)
	print_line(1, 1, termbox.ColorBlack, bar_color, fmt.Sprintf("%v tweets remaining. [%v/%v]", remaining, position, len(tweet_list)-1))

	if len(tweet_list) == 0 {
		return
	}

	// tweet
	tweet := tweet_list[position]
	print_line(1, 2, termbox.ColorMagenta, termbox.ColorBlack, tweet.ScreenName)
	print_line(21, 2, termbox.ColorGreen, termbox.ColorBlack, tweet.UserName)
	print_wordwrap(tweet.Text, 50, 3)
}

func main () {
	flag.Parse()

	stream := make(chan []byte, 1000)
	key := make(chan termbox.Key)
	done := make(chan bool)

	// log.Printf("[config_filename] %#v", *config_filename)
	// httpstream.SetLogger(log.New(os.Stdout, "", log.Ltime|log.Lshortfile), "debug")

	// configuration JSON
	config_file, err := os.Open(*config_filename)
	defer config_file.Close()

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

	client.SetMaxWait(5)

	client.User(done)
	// client.Sample(done)


	go func () {
		for {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				key <- ev.Key
			}
		}
	}()

	drawScreen()

	/* counter := 0
	limit := 2 */

loop:
	for {
		select {
		case k := <-key:
			switch k {
			case termbox.KeyEsc:
				break loop

			case termbox.KeySpace:
				if position < len(tweet_list)-1  {
					position++
					drawScreen()
				}
			}

		case b := <-stream:
			/* if counter >= limit {
				break;
			}

			counter++ */

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

		case <-done:
			termbox.Close()
			log.Print("Client lost connnection.")
		}
	}
}
