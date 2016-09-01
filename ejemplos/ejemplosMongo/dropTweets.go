package main

// twitter oauth

import (
	"flag"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
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
	c := session.DB("test").C("people")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
			// Reading the tweets stored in the BD
			err = c.DropCollection()
			if err != nil { 
				panic(err) 
			} 
}