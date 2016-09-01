package main

import (
	"fmt"
	"flag"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"twittertypes"
)

var (
	maxCt          *int    = flag.Int("maxct", 2, "Max # of messages")
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
	c := session.DB("timelines").C("coordinates")
	var tweet twittertypes.Tweet
	iter := c.Find(nil).Iter()
	contador := 0
	for iter.Next(&tweet) {
		//fmt.Println(tweet.Entities.User_mentions)
		if len(tweet.Geo.Coordinates) > 1 {
				contador++
				fmt.Println(tweet.Geo.Coordinates)
		}

	}
	fmt.Println(contador)

	
	return

}