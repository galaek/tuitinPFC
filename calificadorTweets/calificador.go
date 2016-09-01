package main

import (
	"fmt"
	"labix.org/v2/mgo"			// MongoDB Driver 
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"os/exec"
	"os"
	///"twittertypes"
)

type RatedText struct {
	Text 			string
	Rate			int
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
	//r := session.DB("twitter").C("becasep")
	r := session.DB("f1").C("canada")
	w := session.DB("sentimientos").C("calificaciones")
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
		// Reading the tweets stored in the BD
		//var text []Texto
		//var tweets []twittertypes.Tweet
		var tweetCalificado RatedText
		var texto struct{Id bson.ObjectId "_id"
						 Text string}
		var calificacion int = 100
		iter := r.Find(nil).Select(bson.M{"_id:":1, "text": 1}).Iter()
		if err != nil { 
			panic(err) 
		}

		for iter.Next(&texto) {
			// Mostramos el texto y capturamos la opcion de teclado

			for (calificacion != -1) && ((calificacion < 0) || (calificacion > 5)) {
				cmd := exec.Command("cmd", "/c", "cls")
				cmd.Stdout = os.Stdout
				cmd.Run()
				fmt.Println("-----------------------------------------------------------------")
				fmt.Println("Califica de 1 a 5 los tweets (1 muy malo, 5 muy bueno):")
				fmt.Println("-----------------------------------------------------------------")
				fmt.Println("Texto:")
				//fmt.Println("Id:",texto.Id)
				fmt.Println(texto.Text)
				fmt.Println("-----------------------------------------------------------------")
				fmt.Printf("Introduzca la calificacion (-1 para salir, 0 para saltar): ")
				fmt.Scan(&calificacion)
			}
			switch calificacion {
				case -1:

				
					return
				case 0: // Skip Tweet
					// Esto se hace una vez analizado el tweet
					//r.Remove(bson.M{"_id": texto.Id})
				default:
					tweetCalificado.Text = texto.Text
					tweetCalificado.Rate = calificacion
					err = w.Insert(tweetCalificado) 
					if err != nil { 
						panic(err) 
					}
					// Esto se hace una vez analizado el tweet
					//r.Remove(bson.M{"_id": texto.Id})
					
			}
			calificacion = 100

		}
		
		if err := iter.Close(); err != nil {
			return
		}
		fmt.Println("Done")

}