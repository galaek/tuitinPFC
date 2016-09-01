package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	//"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type User struct {
	Lang                         string
	Verified                     bool
	Followers_count              int
	Location                     string
	Screen_name                  string
	Following                    bool
	Friends_count                int
	Profile_background_color     string
	Favourites_count             int
	Description                  string
	Notifications                string
	Profile_text_color           string
	Url                          string
	Time_zone                    string
	Statuses_count               int
	Profile_link_color           string
	Geo_enabled                  bool
	Profile_background_image_url string
	Protected                    bool
	Contributors_enabled         bool
	Profile_sidebar_fill_color   string
	Name                         string
	Profile_background_tile      string
	Created_at                   string
	Profile_image_url            string
	Id                           int64
	Utc_offset                   int
	Profile_sidebar_border_color string
}

type Request struct {
	User       string
	Timestamp  time.Time
	Smartpoint string
	City       string
	Msg        string
	Option     string
	//location: {type:"Point", coordinates:[]},
	Status string
}

type Entities struct {
	Hashtags []struct {
		Indices []int
		Text    string
	}
	Urls []struct {
		Indices      []int
		Url          string
		Display_url  string
		Expanded_url string
	}
	User_mentions []struct {
		Name        string
		Indices     []int
		Screen_name string
		Id          int64
		Id_str      string
	}
	Media []struct {
		Id              int64
		Id_str          string
		Media_url       string
		Media_url_https string
		Url             string
		Display_url     string
		Expanded_url    string
		Sizes           MediaSizes
		Type            string
		Indices         []int
	}
}
type MediaSizes struct {
	Medium MediaSize
	Thumb  MediaSize
	Small  MediaSize
	Large  MediaSize
}

type MediaSize struct {
	W      int
	H      int
	Resize string
}

type Tweet struct {
	Text                    string
	Truncated               bool
	Geo                     string
	In_reply_to_screen_name string
	Favorited               bool
	Source                  string
	Contributors            string
	In_reply_to_status_id   string
	In_reply_to_user_id     int64
	Id                      int64
	Created_at              string
	User                    *User
	Entities                *Entities
}

type Tarea struct {
	Command string
	Text    string
}
type LedTarea struct {
	Command string
}

var (
	consumerKey    *string = flag.String("ck", "BOWIAZSb7FpUidsCYqAtQ", "Consumer Key")
	consumerSecret *string = flag.String("cs", "3KsKhC9QWGDnBPBoy52cmHjFsbR65VNdKFRTfCXVGY", "Consumer Secret")
	ot             *string = flag.String("ot", "2287121436-77t6FJzCOfGnnRMXYdC8UMCyxpxM3SogM5ammSJ", "Oauth Token")
	osec           *string = flag.String("os", "I5rCaWrMdRdhgyxRlsRAgMq6L1Cm6Re547sfoAyBbjHSq", "OAuthTokenSecret")
	logLevel       *string = flag.String("logging", "info", "Which log level: [debug,info,warn,error,fatal]")
	search         *string = flag.String("search", "SmartZity", "keywords to search for, comma delimtted")
)

func ledworker(pool chan LedTarea) {
	fmt.Print("Led working...")
	var mt LedTarea
	buf := `{
    "commandML":"<paid:command name=\"SET\"><paid:cmdParam name=\"FreeText\"><swe:Text><swe:value>FIZCOMMAND %s </swe:value></swe:Text></paid:cmdParam></paid:command>"
}`

	for {
		mt = <-pool
		if mt.Command == "RED" {
			println("Serving Led RED")
			cadena := "99-0-0"
			mbuf := fmt.Sprintf(buf, cadena)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			//enviar el commando con el texto
		} else if mt.Command == "BLUE" {
			println("Serving Led BLUE")
			//enviar el commando con la foto
			cadena := "0-99-0"
			mbuf := fmt.Sprintf(buf, cadena)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
		} else if mt.Command == "GREEN" {
			println("Serving Led GREEN")
			//enviar el commando con la foto
			cadena := "0-0-99"
			mbuf := fmt.Sprintf(buf, cadena)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
		} else if mt.Command == "RAINBOW" {
			println("Serving Led RAINBOW")
			cadena := "0-99-0"
			mbuf := fmt.Sprintf(buf, cadena)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "0-99-99"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "0-0-99"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "99-0-99"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "99-0-0"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "99-99-0"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
			cadena = "0-99-0"
			mbuf = fmt.Sprintf(buf, cadena)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/RGBS:C3:42:24:0005/command",
				"application/json", body)
			time.Sleep(time.Second)
		}

		time.Sleep(10 * time.Second)
	}
	fmt.Println("done")
}

func worker(pool chan Tarea) {
	fmt.Print("working...")
	var mt Tarea
	n := 0
	buf := `{
    "commandML":"<paid:command name=\"SET\"><paid:cmdParam name=\"FreeText\"><swe:Text><swe:value>FIZCOMMAND %s </swe:value></swe:Text></paid:cmdParam></paid:command>"
}`

	for {
		mt = <-pool
		n++
		if mt.Command == "TEXT" {

			cadena := "TXT-" + strings.Replace(mt.Text, " ", "_", -1)
			mbuf := fmt.Sprintf(buf, cadena)
			println("Serving Text: ", mbuf)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
			//enviar el commando con el texto
		} else {
			//enviar el commando con la foto
			cadena := "TXT-CARGANDO_IMAGEN"
			mbuf := fmt.Sprintf(buf, cadena)
			println("Serving Text: ", mbuf)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
			cadena = "IMG-" + mt.Text
			mbuf = fmt.Sprintf(buf, cadena)
			println("Serving Imagen: ", mbuf)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
		}

		time.Sleep(10 * time.Second)

		if n == 3 {
			cadena := "TXT-CARGANDO_IMAGEN"
			mbuf := fmt.Sprintf(buf, cadena)
			println("Serving Text: ", mbuf)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
			cadena = "IMG-" + "http://s7.postimg.org/ctymbsy8r/smartz.jpg"
			mbuf = fmt.Sprintf(buf, cadena)
			println("Serving Promo smartz:", mbuf)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
		} else if n == 6 {
			cadena := "TXT-CARGANDO_IMAGEN"
			mbuf := fmt.Sprintf(buf, cadena)
			println("Serving Text: ", mbuf)
			body := bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
			cadena = "IMG-" + "http://s7.postimg.org/agbmnodi3/powered.jpg"
			mbuf = fmt.Sprintf(buf, cadena)
			println("Serving Promo powered: ", mbuf)
			body = bytes.NewBuffer([]byte(mbuf))
			http.Post("http://130.206.80.44:5371/m2m/v2/services/HACKBRAZIL/devices/LCD:C3:42:24:0002/command",
				"application/json", body)
		} else if n == 7 {
			n = 0
		}
		time.Sleep(10 * time.Second)
	}
	fmt.Println("done")
}

func main() {

	flag.Parse()
	httpstream.SetLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile), *logLevel)

	// make a go channel for sending from listener to processor
	// we buffer it, to help ensure we aren't backing up twitter or else they cut us off
	stream := make(chan []byte, 1000)
	done := make(chan bool)
	tareasLED := make(chan LedTarea, 1000)
	tareasLCD := make(chan Tarea, 1000)
	go ledworker(tareasLED)
	go worker(tareasLCD)

	httpstream.OauthCon = oauth.NewConsumer(
		*consumerKey,
		*consumerSecret,	
		oauth.ServiceProvider{
			RequestTokenUrl:   "http://api.twitter.com/oauth/request_token",
			AuthorizeTokenUrl: "https://api.twitter.com/oauth/authorize",
			AccessTokenUrl:    "https://api.twitter.com/oauth/access_token",
		})
	
	at := oauth.AccessToken{
		Token:  *ot,
		Secret: *osec,
	}
	// the stream listener effectively operates in one "thread"/goroutine
	// as the httpstream Client processes inside a go routine it opens
	// That includes the handler func we pass in here
	client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
		stream <- line
		// although you can do heavy lifting here, it means you are doing all
		// your work in the same thread as the http streaming/listener
		// by using a go channel, you can send the work to a
		// different thread/goroutine
	}))
	
	var keywords []string
	if search != nil && len(*search) > 0 {
		keywords = strings.Split(*search, ",")
	}
	err := client.Filter(nil, keywords, nil, nil, false, done)
	if err != nil {
		httpstream.Log(httpstream.ERROR, err.Error())
	} else {
		go func() {
			var tweet Tweet
			for tw := range stream {
				json.Unmarshal(tw, &tweet)
				fmt.Println(tweet.Text)
				fmt.Println(tweet.User.Screen_name)
				fmt.Println(tweet.User.Id)
				/*out, err := time.Parse(time.RubyDate, tweet.Created_at)
				if err != nil {
					panic("Could not parse time")
				}
				*/if len(tweet.Entities.Hashtags) == 5 {

					if tweet.Entities.Hashtags[4].Text == "TEXT" && tweet.Entities.Hashtags[3].Text == "SP02" {
						tarea := Tarea{tweet.Entities.Hashtags[4].Text, tweet.Text[0:tweet.Entities.Hashtags[0].Indices[0]]}
						tareasLCD <- tarea
						//estamos en sao paulo

					} else if tweet.Entities.Hashtags[4].Text == "IMG" && tweet.Entities.Hashtags[3].Text == "SP02" {
						tarea := Tarea{tweet.Entities.Hashtags[4].Text, tweet.Entities.Media[0].Media_url}
						tareasLCD <- tarea
					} else if tweet.Entities.Hashtags[3].Text == "SP01" {
						//si es un led
						tarea := LedTarea{tweet.Entities.Hashtags[4].Text}
						tareasLED <- tarea
					} else {
						/*
							session, err := mgo.Dial("localhost:27017")
							if err != nil {
								panic(err)
							}
							defer session.Close()
							c := session.DB("smartzity").C("requests")
							var r Request
							r.User = tweet.User.Screen_name
							r.Timestamp = out
							r.City = tweet.Entities.Hashtags[2].Text
							r.Msg = tweet.Text
							r.Option = tweet.Entities.Hashtags[4].Text
							r.Smartpoint = tweet.Entities.Hashtags[3].Text
							r.Status = "NEW"
							err = c.Insert(r)
						*/
						println("Error de codigo")
						println(tweet.Entities.Hashtags[0].Text)
						println(tweet.Entities.Hashtags[1].Text)
						println(tweet.Entities.Hashtags[2].Text)
						println(tweet.Entities.Hashtags[3].Text)
						println(tweet.Entities.Hashtags[4].Text)
					}
					if err != nil {
						panic(err)
					}
				}
			}
		}()
		_ = <-done
	}

}
