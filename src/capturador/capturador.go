package main

import (
	"fmt"
	"flag"
	"encoding/json"
	"github.com/araddon/httpstream"
	"github.com/mrjones/oauth"
	"log"
	"os"
	"strconv"
	"strings"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"twittertypes"
	"time"
)

var (
	maxCt          *int    = flag.Int("maxct", 1000000000, "Max # of messages")
	user           *string = flag.String("user", "", "twitter username")
	consumerKey    *string = flag.String("ck", "Cij69U0URS7d0dMNkw9p2nfJp", "Consumer Key")
	consumerSecret *string = flag.String("cs", "gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z", "Consumer Secret")
	// ot             *string = flag.String("ot", "2425498608-LuXL0xwG2tpHJzzBJN4XTIP3W2RhzoMYukos57i", "Oauth Token")
	// osec           *string = flag.String("os", "4UdIltWpL8S69k3gcYDFfGcjiqeIp2aUjca0UxALIM2BQ", "OAuthTokenSecret")
	ot             *string = flag.String("ot", "2606134326-W4aZaT19VidhL4onwTwl4C9Cg5CrivI0uPB9njA", "Oauth Token")
	osec           *string = flag.String("os", "3AaNzQN6VuF80JHiyhe411LW1SOkeOO36P9XILlu7kaX7", "OAuthTokenSecret")
	logLevel       *string = flag.String("logging", "info", "Which log level: [debug,info,warn,error,fatal]")
	search         *string = flag.String("search", "#Mundial2014,#Brasil2014,#Brazil2014,#WorldCup", "keywords to search for, comma delimtted")
	users          *string = flag.String("users", "", "list of twitter userids to filter for, comma delimtted")
	collection     *string = flag.String("collection", "mundialBrasil", "name of the collection where we store the Tweets")
	mongoip		   *string = flag.String("mongoip", "130.206.83.133", "IP of the machine with the MongoDB")
	duration	   *string = flag.String("duration", "150m", "Interval in minutes where we capture tweets")
)

func main() {

	flag.Parse()
	httpstream.SetLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile), *logLevel)

	// Connectin to the local mongo database
	session, err := mgo.Dial(*mongoip) 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	t:= time.Now()
	collectionName := fmt.Sprintf("%d-%d-%d_%d:%d_%s", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), *collection)
	fmt.Println("Collection:", collectionName)
	// Creating DB "twitter" and Collection "tweets"
	c := session.DB("mundial").C(collectionName)
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 

	// make a go channel for sending from listener to processor
	// we buffer it, to help ensure we aren't backing up twitter or else they cut us off
	stream := make(chan []byte, 1000)
	done := make(chan bool)

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

	// Creamos el cliente OAuth
	client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) { stream <- line }))
	fmt.Println("Cliente OAuth creado")
	// find list of userids we are going to search for
	userIds := make([]int64, 0)
	for _, userId := range strings.Split(*users, ",") {
		if id, err := strconv.ParseInt(userId, 10, 64); err == nil {
			userIds = append(userIds, id)
		}
	}
	var keywords []string
	if search != nil && len(*search) > 0 {
		keywords = strings.Split(*search, ",")
	}
	//err := client.Filter(userIds, keywords, []string{"es"}, nil, false, done)
	//var locats []string
	// locats := make([]string, 4)
	// locats[0] = "-180"
	// locats[1] = "-90"
	// locats[2] = "180"
	// locats[3] = "90"
	//err = client.Filter(userIds, keywords, nil, locats, false, done)
	err = client.Filter(userIds, keywords, nil, nil, false, done)
	//err := client.Filter(nil, keywords, nil, nil, false, done)
	if err != nil {
		httpstream.Log(httpstream.ERROR, err.Error())
	} else {
		fmt.Println("Filter realizado, esperando datos")
		durr, err := time.ParseDuration(*duration)
		if err != nil {
			panic(err)
		}
		timeNow := time.Now()
		timeEnd := time.Now().Add(durr)
		fmt.Println("Start:", timeNow)
		fmt.Println("End:", timeEnd)
		fmt.Println("Duration:", durr)
		//fmt.Println("Epoch start:", timeNow.Unix())
		//fmt.Println("Epoch end:", timeEnd.Unix())
		
		go func() {
			// while this could be in a different "thread(s)"
			ct := 0
			//var tweet twittertypes.Tweet
			for tw := range stream {
				tweet := new(twittertypes.TweetEncuesta)
				//println("Tweet", ct, "capturado")
				json.Unmarshal(tw, &tweet)
				tweetstamp, err := time.Parse(time.RubyDate, tweet.Created_at)
				tweet.Epoch = tweetstamp.Unix()
				//if ct >= *maxCt {
				if timeEnd.Unix() < tweet.Epoch {
					fmt.Println("Num tweets captured:", ct)
					done <- true
				} else {
					err = c.Insert(tweet) 
					if err != nil { 
						panic(err) 
					} 
					ct++
				}
				// heavy lifting
			}			
		}()
		_ = <-done
	}
}