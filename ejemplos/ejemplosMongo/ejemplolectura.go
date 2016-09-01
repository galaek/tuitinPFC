package main

import (
	"fmt"
	"sort"
	"strings"
	"flag"
	"github.com/araddon/httpstream"
	"labix.org/v2/mgo"			// MongoDB Driver 
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
)

var (
	maxCt          *int    = flag.Int("maxct", 2, "Max # of messages")
)

func addWord(m map[string]int, palabra string) {
	if palabra == "" {
		return
	}
	valor, existe := m[palabra]
	if existe == false {
		m[palabra] = 1
	} else {
		m[palabra] = valor+1
	}
}

type Par struct {
	Palabra string
	Valor int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type listaPalabras []Par
func (p listaPalabras) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p listaPalabras) Len() int { return len(p) }
func (p listaPalabras) Less(i, j int) bool { return p[i].Valor > p[j].Valor } // en este caso implementado >

// A function to turn a map into a PairList, then sort and return it. 
func sortMapByValue(m map[string]int) listaPalabras {
   p := make(listaPalabras, len(m))
   i := 0
   for k, v := range m {
      p[i] = Par{k, v}
	  i++
   }
   sort.Sort(p)
   return p
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
	c := session.DB("twitter").C("ejemplo1")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
		var tweets []httpstream.Tweet
		err = c.Find(bson.M{"text": bson.M{"$exists": true}}).All(&tweets)
		if err != nil { 
			panic(err) 
		} 
		
		// AQUI EL ANALISIS DE LOS TWEETS.		
		fmt.Println("Num Tweets: ", len(tweets))
		for i := 0; i < len(tweets) ; i++ {
			fmt.Println("Texto ", i, ": ", tweets[i].Text)
		}
			
		palabras := make(map[string]int)	
		var texto string
		var resultado []string
		var palabrasOrdenadas listaPalabras
		for i:= 0; i < len(tweets); i++ {
			texto = tweets[i].Text
			resultado = textoApalabras(texto)
			// Metemos las palabras en un mapa
			for j:=0; j< len(resultado); j++ {
				addWord(palabras, resultado[j])
			}
			palabrasOrdenadas = sortMapByValue(palabras)
		}
		
		fmt.Println(palabrasOrdenadas)
		for i:=0; i< len(palabrasOrdenadas); i++ {
			fmt.Println(palabrasOrdenadas[i].Valor, palabrasOrdenadas[i].Palabra)
		}
		
			
}

func textoApalabras (t string) []string {
	t = strings.Replace(t, ".", " ", -1)
	t = strings.Replace(t, ",", " ", -1)
	t = strings.Replace(t, ":", " ", -1)
	t = strings.Replace(t, "!", " ", -1)
	t = strings.Replace(t, "?", " ", -1)
	t = strings.ToLower(t)
	return strings.Split(t, " ")
} 
