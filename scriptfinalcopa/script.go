package main

import (
	"fmt"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"twittertypes"
	"time"
)

func main() {

	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	// Creating DB "twitter" and Collection "tweets"
	c := session.DB("twitter").C("finalcopa")
	g := session.DB("twitter").C("finalcopafixed2")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
	var tweets twittertypes.Tweet
	iter := c.Find(nil).Iter()
	contador := 0;
	for iter.Next(&tweets) {
		tweetstamp, err := time.Parse(time.RubyDate, tweets.Created_at)
		if err != nil { 
			panic(err) 
		} 
		tweets.Epoch = tweetstamp.Unix()
		err = g.Insert(tweets) 
		if err != nil { 
			panic(err) 
		} 
		contador++
	}
	if err := iter.Close(); err != nil {
		fmt.Println("Ha habido error")
	}
	fmt.Println("Tweets contados:", contador)			
}