package main

import (
	//"fmt"
	"flag"
	"labix.org/v2/mgo"			// MongoDB Driver 
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	//"twittertypes"
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
	r := session.DB("final").C("finalchampions")
	w := session.DB("sentimientos").C("textos")
	
	var texto struct{Text string}
	iter := r.Find(nil).Select(bson.M{"text": 1}).Iter()
	if err != nil { 
		panic(err) 
	}
	for iter.Next(&texto) {
		//fmt.Println(texto.Text)
		err = w.Insert(texto) 
		if err != nil { 
			panic(err) 
		}
	}
	
	return
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
		// var tweets []twittertypes.Tweet
		// err = c.Find(nil).All(&tweets)
		// if err != nil { 
			// panic(err) 
		// } 
		
		// AQUI EL ANALISIS DE LOS TWEETS.		
		// for i:=0; i<len(tweets); i++ {
			// fmt.Println(tweets[i].Text)
		// }
		// fmt.Println("Num tweetz:", len(tweets))
}