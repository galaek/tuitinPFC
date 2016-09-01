package main

import (
	"fmt"
	"flag"
	"labix.org/v2/mgo"			// MongoDB Driver 
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
)

var (
	maxCt          *int    = flag.Int("maxct", 2, "Max # of messages")
)

type Par struct {
	Palabra string
	Valor int
}

func main() {

	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	// Creating DB "twitter" and Collection "tweets"
	c := session.DB("twitter").C("resultado1")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
		var listaPalabras []Par
		err = c.Find(bson.M{"palabra": bson.M{"$exists": true}}).All(&listaPalabras)
		if err != nil { 
			panic(err) 
		} 
		
		// AQUI EL ANALISIS DE LOS TWEETS.		
		fmt.Println("Num Palabras Distintas: ", len(listaPalabras))
		for i := 0; i < len(listaPalabras) ; i++ {
			fmt.Println("Num Veces:", listaPalabras[i].Valor, "Palabra:", listaPalabras[i].Palabra)
		}
}