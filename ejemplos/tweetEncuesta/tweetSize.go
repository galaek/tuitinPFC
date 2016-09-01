package main

// Using application-only auth.

import (
	"fmt"
	"twittertypes"
	//"strings"
	"labix.org/v2/mgo"	
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"profiles"
	//"trendings"
	"unsafe"

)

func main() {
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	c := session.DB("testsize").C("tweetsize")
	
	profiles.GetTimelineToDB(*c, "manupruebas", 1) 
	var tweets []twittertypes.TweetEncuesta
	err = c.Find(bson.M{"text": bson.M{"$exists": true}}).All(&tweets)
	if err != nil { 
		panic(err) 
	}
	
	//todosLosTweets := trendings.GetTweetsAll(*c)
	fmt.Println(len(tweets))
	fmt.Println(tweets[0])
	fmt.Println("Un tweet son:", unsafe.Sizeof(tweets[0]), "bytes")
	fmt.Println("Un user son:", unsafe.Sizeof(*tweets[0].User), "bytes")
	fmt.Println("La estructura total son:", unsafe.Sizeof(tweets[0])+unsafe.Sizeof(*tweets[0].User), "bytes")
	err = c.DropCollection()
	if err != nil { 
		panic(err) 
	}	
	
}