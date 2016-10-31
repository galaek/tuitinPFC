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
	ot             *string = flag.String("ot", "2425498608-c42EqgkqWq7RsBrxknqopf1zxAdvpqJzCwGwDg4", "Oauth Token")
	osec           *string = flag.String("os", "21QYsovag4ZVle3B3ILEVx0fDXP6BMWUfiTDxYT0p9aHW", "OAuthTokenSecret")
	logLevel       *string = flag.String("logging", "info", "Which log level: [debug,info,warn,error,fatal]")
	search         *string = flag.String("search", "#WPTZaragozaOpen", "keywords to search for, comma delimtted")
	users          *string = flag.String("users", "", "list of twitter userids to filter for, comma delimtted")
	collection     *string = flag.String("collection", "zaragoza", "name of the collection where we store the Tweets")
)

func main() {

	flag.Parse()
	httpstream.SetLogger(log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile), *logLevel)

	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	t:= time.Now()
	collectionName := fmt.Sprintf("%d-%d-%d_%d:%d_%s", t.Year(), int(t.Month()), t.Day(), t.Hour(), t.Minute(), *collection)
	fmt.Println("Collection:", collectionName)
	// Creating DB "twitter" and Collection "tweets"
	c := session.DB("WPT").C(collectionName)
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
		go func() {
			// while this could be in a different "thread(s)"
			ct := 0
			//var tweet twittertypes.Tweet
			for tw := range stream {
				//println("Tweet", ct, "capturado")
				tweet := new(twittertypes.Tweet)
				json.Unmarshal(tw, &tweet)
				tweetstamp, err := time.Parse(time.RubyDate, tweet.Created_at)
				tweet.Epoch = tweetstamp.Unix()
				err = c.Insert(tweet) 
				if err != nil { 
					panic(err) 
				} 
				// heavy lifting
				ct++
				if ct >= *maxCt {
					done <- true
				}
			}			
		}()
		_ = <-done
	}
}