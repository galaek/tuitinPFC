package main

import (
	"fmt"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"twittertypes"
	//"time"
	"math"
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
	c := session.DB("twitter").C("finalcopafixed")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
	var tweets twittertypes.Tweet
	//iter := c.Find(bson.M{"text": bson.M{"$exists": true}}).Iter()
	iter := c.Find(nil).Iter()
	contador := 0;
	for iter.Next(&tweets) {
	//for i:=0; i<100000; i++ {
	//	iter.Next(&tweets)
		if math.Mod(float64(contador),10000) == 0 {
			fmt.Println("Created at:", tweets.Created_at, "Epoch:", tweets.Epoch)
		}
		contador++
	}
	fmt.Println("Created at:", tweets.Created_at, "Epoch:", tweets.Epoch)
	if err := iter.Close(); err != nil {
		fmt.Println("Ha habido error")
	}
	fmt.Println("Tweets contados:", contador)			
}