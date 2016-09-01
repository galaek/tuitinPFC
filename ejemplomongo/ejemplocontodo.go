package main

// twitter oauth

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
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
)

var (
	maxCt          *int    = flag.Int("maxct", 2, "Max # of messages")
	user           *string = flag.String("user", "", "twitter username")
	consumerKey    *string = flag.String("ck", "Cij69U0URS7d0dMNkw9p2nfJp", "Consumer Key")
	consumerSecret *string = flag.String("cs", "gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z", "Consumer Secret")
	ot             *string = flag.String("ot", "2425498608-LuXL0xwG2tpHJzzBJN4XTIP3W2RhzoMYukos57i", "Oauth Token")
	osec           *string = flag.String("os", "4UdIltWpL8S69k3gcYDFfGcjiqeIp2aUjca0UxALIM2BQ", "OAuthTokenSecret")
	logLevel       *string = flag.String("logging", "info", "Which log level: [debug,info,warn,error,fatal]")
	search         *string = flag.String("search", "testingDeViernes", "keywords to search for, comma delimtted")
	users          *string = flag.String("users", "", "list of twitter userids to filter for, comma delimtted")
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
	// Creating DB "twitter" and Collection "tweets"
	c := session.DB("twitter").C("tweets")
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
	// the stream listener effectively operates in one "thread"/goroutine
	// as the httpstream Client processes inside a go routine it opens
	// That includes the handler func we pass in here
/*	client := httpstream.NewOAuthClient(&at, httpstream.OnlyTweetsFilter(func(line []byte) {
		stream <- line
		// although you can do heavy lifting here, it means you are doing all
		// your work in the same thread as the http streaming/listener
		// by using a go channel, you can send the work to a
		// different thread/goroutine
	})) */
	
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
	err = client.Filter(userIds, keywords, nil, nil, false, done)
	//err := client.Filter(nil, keywords, nil, nil, false, done)
	if err != nil {
		httpstream.Log(httpstream.ERROR, err.Error())
	} else {
		fmt.Println("Filter realizado, esperando datos")
		go func() {
			// while this could be in a different "thread(s)"
			ct := 0
			var tweet httpstream.Tweet
			for tw := range stream {
				println("Guardamos Tweet")
				json.Unmarshal(tw, &tweet)
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
		
			// Reading the tweets stored in the BD
			var tweets []httpstream.Tweet
			err = c.Find(bson.M{"text": bson.M{"$exists": true}}).All(&tweets)
			//err = c.Find(bson.M{"name": "Ale"}).All(&persona) 
			//err = c.Find(bson.M{"name": "Ale"}).One(&result)
			if err != nil { 
				panic(err) 
			} 
			fmt.Println("Num Tweets: ", len(tweets))
			for i := 0; i < len(tweets) ; i++ {
				fmt.Println("Texto ", i, ": ", tweets[i].Text)
			}
		
	}

}
