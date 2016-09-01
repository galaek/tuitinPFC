package main

import (
	"fmt"
	//"github.com/araddon/httpstream"
	"labix.org/v2/mgo"			// MongoDB Driver 
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"time"
	"twittertypes"
	"net/http"
	"html/template"
	"trendings"
	"profiles"
)

var todosLosTweets []twittertypes.Tweet
var listaPalabras trendings.WordsList
var mergedList trendings.WordsList
var listaDoblesPalabras trendings.WordsList
var tweetsFiltrados []twittertypes.Tweet
var listaPalabrasKeyword trendings.WordsList
var listaDoblesPalabrasKeyword trendings.WordsList
var listaPalabrasRango1 trendings.WordsList
var listaPalabrasRango2 trendings.WordsList
var listaDoblesPalabrasRango1 trendings.WordsList
var listaDoblesPalabrasRango2 trendings.WordsList
var mapaResultado map[int64]trendings.VerifiedUser
var listaResultado trendings.ListaFavs



func main() {

	// Topics testing
	//InterestsTesting()
	//return
	// Testing the new TweetToText functions
	//TweetToTextTesting()
	//return
	// Testing the interval functions
	//IntervalsTesting()
	//return
	// Testing the database and tweet parsing functions
	DataBaseTestAndParsing()
	// Testing GetTweetsKeyword
	//TweetsKeywordTest()
	//return
	// Testing the Verified Tweets functions
	VerifiedTesting()
	
	// Testing RT and Fav
	//RtAndFavTesting()
	
	//return
	
	fmt.Println("Servidor iniciado, esperando peticiones")

	// Aqui vamos a hacer un ejemplo de mostrar cosas en bootstrap
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/golang/ejemplomongo/contarpalabras/web/"))))
	http.HandleFunc("/view/", myHandler)
    http.ListenAndServe(":8080", nil)
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, "Hola View!!!")
	type resultados struct {
		Tweets 				[]twittertypes.Tweet
		Lista				trendings.WordsList
		ListaDouble			trendings.WordsList
		ShortList 			[7]trendings.Pair
		ShortDoubleList 	[7]trendings.Pair
		ShortListKeyword 	[7]trendings.Pair
		ShortDoubleListKeyword 	[7]trendings.Pair
		ShortListRange1 	[7]trendings.Pair
		ShortListRange2		[7]trendings.Pair
		MergedList			[7]trendings.Pair
		VerifiedsMap		map[int64]trendings.VerifiedUser
		VerifiedsList		trendings.ListaFavs
	}
	
	var res resultados
	res.Tweets = todosLosTweets
	//res.Tweets = tweetsFiltrados
	res.Lista = listaPalabras
	res.ListaDouble = listaDoblesPalabras
	for i:=0; i<7; i++ {
		res.ShortList[i] = res.Lista[i]
	}
	for i:=0; i<7; i++ {
		res.ShortDoubleList[i] = res.ListaDouble[i]
		fmt.Printf("-->%s<--\n", res.ShortDoubleList[i].Word)
	}
	for i:=0; i<7; i++ {
		res.ShortListKeyword[i] = listaPalabrasKeyword[i]
	}
	for i:=0; i<7; i++ {
		res.ShortDoubleListKeyword[i] = listaDoblesPalabrasKeyword[i]
	}
	for i:=0; i<7; i++ {
		res.MergedList[i] = mergedList[i]
	}
	for i:=0; i<7; i++ {
		res.ShortListRange1[i] = listaPalabrasRango1[i]
		res.ShortListRange2[i] = listaPalabrasRango2[i]
		
	}
	res.VerifiedsMap = mapaResultado
	res.VerifiedsList = listaResultado
	t, _ := template.ParseFiles("web/starter-template/index.html")
    t.Execute(w, res)
}

// Testing Functions

func InterestsTesting () {
	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	
	var twitterProfile string
	// twitterProfile = "nexocargo"
	// twitterProfile = "enrimr"
	// twitterProfile = "548017" // wences
	// twitterProfile = "alvjimcesar"
	// twitterProfile = "daniel_prol"
	// twitterProfile = "saracaan"
	// twitterProfile = "irenetaule"
	// twitterProfile = "la_terminal_"
	// twitterProfile = "sonosara_mk"
	// twitterProfile = "alberto_s_h"
	//twitterProfile = "srbuhito"
	 twitterProfile = "steamkhemia"
	c := session.DB("timelines").C(twitterProfile)
	profiles.GetTimelineToDB(*c, twitterProfile, 2000) 
	todosLosTweets := trendings.GetTweetsAll(*c)
	listaPalabras, listaDoblesPalabras := trendings.ParseTweetsAll(todosLosTweets, "")
	
	resultados := make([]int, len(profiles.TopicsList))
	profiles.LoadTopics()
	for i:=0; i<len(listaPalabras); i++ {
		for j:=0; j<len(profiles.TopicsList); j++ {
			value, exists := profiles.TopicsDB[profiles.TopicsList[j]][listaPalabras[i].Word]
			if exists == true {
				resultados[j] = resultados[j] + (listaPalabras[i].Value * value)
				fmt.Println(listaPalabras[i].Word, listaPalabras[i].Value )
			}	
		}
	}
	for i:=0; i<len(listaDoblesPalabras); i++ {
		for j:=0; j<len(profiles.TopicsList); j++ {
			value, exists := profiles.TopicsDB[profiles.TopicsList[j]][listaDoblesPalabras[i].Word]
			if exists == true {
				resultados[j] = resultados[j] + (listaDoblesPalabras[i].Value * value)
				fmt.Println(listaDoblesPalabras[i].Word, listaDoblesPalabras[i].Value )
			}	
		}
	}

	fmt.Println("")
	for i:=0; i<len(resultados); i++ {
		fmt.Println(profiles.TopicsList[i], ":", resultados[i])
	}

	err = c.DropCollection()
	if err != nil { 
		panic(err) 
	}	
	return 
}

func TweetToTextTesting() {
	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close()
	
	a := session.DB("timelines").C("conurls")
	//profiles.GetTimelineToDB(*a, "manupruebas", 5) // 2, 9, 25, 30, 40....
	var tweets []twittertypes.Tweet
	err = a.Find(nil).All(&tweets)
	if err != nil { 
		panic(err) 
	}
	wordsResult, doubleWordsResult := trendings.ParseTweetsAll(tweets, "")
	fmt.Println("Num Words:", len(wordsResult))
	for i:=0; i<len(wordsResult); i++ {
		fmt.Printf("-->%s<--,%d\n", wordsResult[i].Word, wordsResult[i].Value)
	}
	fmt.Println("Num Double Words:", len(doubleWordsResult))
	for i:=0; i<len(doubleWordsResult); i++ {
		fmt.Printf("-->%s<--,%d\n", doubleWordsResult[i].Word, doubleWordsResult[i].Value)
	}
	return
}
func RtAndFavTesting () {
	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 

	//a := session.DB("twitter").C("becasep")
	//a := session.DB("twitter").C("championstve")
	a := session.DB("final").C("finalchampions")
	var top profiles.Popularity
	top = profiles.GetMostRetweetedAndFavorited(*a)
	fmt.Println("------------------------------------")
	fmt.Println("Total Received:", top.TotalTweets)
	fmt.Println("RtMin", top.MinRt)
	fmt.Println("RtMax", top.MaxRt)
	fmt.Println("FavMin", top.MinFav)
	fmt.Println("FavMax", top.MaxFav)
	fmt.Println("------------------------------------")
	fmt.Println("    RETWEETS")
	fmt.Println("------------------------------------")
	for i:=0; i<5; i++ {
		fmt.Println("Rt", i, "=", top.MostRt[i].Retweet_count, top.MostRt[i].Text)
		//fmt.Println("ID", i, "=", top.MostRt[i].Retweeted_status.Id)
		//fmt.Println("Texto:", top.MostRt[i].Text)
	}
	fmt.Println("------------------------------------")
	fmt.Println("    FAVORITOS")
	fmt.Println("------------------------------------")
	for i:=0; i<5; i++ {
		fmt.Println("Fav", i, "=", top.MostFav[i].Favorite_count, top.MostFav[i].Text)
		//fmt.Println("ID", i, "=", top.MostFav[i].Retweeted_status.Id)
		//fmt.Println("Texto:", top.MostFav[i].Text)
	}
	fmt.Println("------------------------------------")

	return

}



func TweetsKeywordTest() {
	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	
	//a := session.DB("twitter").C("becasep")
	a := session.DB("twitter").C("championstve")
	//a := session.DB("final").C("finalchampions")
	
	tweets := trendings.GetTweetsKeyword(*a, "ramos")
	fmt.Println("Tweets que contienen 'ramos'", len(tweets))
}

func IntervalsTesting() {
	fmt.Println("Intervals Test:")
	start, _ := time.Parse(time.RubyDate, "Tue May 06 09:50:12 +0200 2014")
	end, _ := time.Parse(time.RubyDate, "Wed May 07 05:27:47 +0200 2014")
	//end, err := time.Parse(time.RubyDate, "Tue May 06 15:23:15 +0200 2014")
	fmt.Println("Duracion Total:", end.Sub(start))
	fmt.Println("Originals:")
	fmt.Println(start)
	fmt.Println(end)
	fmt.Println("-------------------------")
	var intervalsTest []trendings.Interval
	intervalsTest = trendings.GetIntervalsN(start, end, 6)
	
	for i:=0; i<len(intervalsTest); i++ {
		fmt.Println(intervalsTest[i].Start)
		fmt.Println(intervalsTest[i].End)
		fmt.Println(intervalsTest[i].End.Sub(intervalsTest[i].Start))
		fmt.Println("---------------------------------")
	}
	
	var dur time.Duration
	dur, _ = time.ParseDuration("3h25m")
	
	var intervalsTest2 []trendings.Interval
	intervalsTest2 = trendings.GetIntervalsLen(start, end, dur)
	fmt.Println("-------------------------")
	fmt.Println("-------------------------")
	fmt.Println("Duracion Total:", end.Sub(start))
	fmt.Println("Originals:")
	fmt.Println(start)
	fmt.Println(end)
	fmt.Println("-------------------------")
	fmt.Println("-------------------------")
	
	for i:=0; i<len(intervalsTest2); i++ {
		fmt.Println(intervalsTest2[i].Start)
		fmt.Println(intervalsTest2[i].End)
		fmt.Println(intervalsTest2[i].End.Sub(intervalsTest2[i].Start))
		fmt.Println("---------------------------------")
	}	
}

func VerifiedTesting () {
	fmt.Println("Verified testing")
	// Connecting to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	
	a := session.DB("twitter").C("championstve")
	//a := session.DB("final").C("finalchampions")
	//func GetTweetsVerified (colec mgo.Collection) []twittertypes.Tweet {
	var tweetsVerified []twittertypes.Tweet
	tweetsVerified = trendings.GetTweetsVerified(*a)
	/*for i:=0; i<len(tweetsVerified); i++ {
		fmt.Println(i, tweetsVerified[i].User.Screen_name)
	} */
	
	mapaResultado = trendings.ListVerifiedUsers(tweetsVerified)
	/*for k := range mapaResultado {
		fmt.Println(mapaResultado[k].Screen_name, mapaResultado[k].Value)
	}*/
	listaResultado = trendings.SortMapByValue2(mapaResultado)
	for j := 0; j<len(listaResultado); j++ {
		fmt.Println(listaResultado[j].User.Screen_name, "-", listaResultado[j].User.Value)
	}
	
}

func DataBaseTestAndParsing () {
	// Connectin to the local mongo database
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	// We defer the session closing so we don't forget later
	defer session.Close() 
	
	//a := session.DB("twitter").C("becasep")
	a := session.DB("twitter").C("championstve")
	//a := session.DB("timelines").C("conurls")	
	
	todosLosTweets = trendings.GetTweetsAll(*a)
	fmt.Println("=============================")
	fmt.Println("Tweets sin filtrar")
	fmt.Println("=============================")
	fmt.Println("Num tweets:", len(todosLosTweets))
	// for i:=0; i<len(todosLosTweets); i++ {
		// fmt.Println(i, todosLosTweets[i].Epoch, todosLosTweets[i].Created_at)
	// }
	
	fmt.Println("============================================")
	fmt.Println("Cuenta de Palabras SIN Keyword sobre TODOS")
	fmt.Println("============================================")
	
	listaPalabras, listaDoblesPalabras = trendings.ParseTweetsAll(todosLosTweets, "#championstve")
	fmt.Println("Num palabras:", len(listaPalabras))
	fmt.Println("Num dobles:", len(listaDoblesPalabras))
	//fmt.Println(listaPalabras)
	//return
	fmt.Println("============================================")
	fmt.Println("Cuenta de Palabras CON Keyword sobre TODOS")
	fmt.Println("============================================")

	listaPalabrasKeyword, listaDoblesPalabrasKeyword = trendings.ParseTweetsKeyword(todosLosTweets, "ramos", "#championstve")
	fmt.Println("Num palabras:", len(listaPalabrasKeyword))
	fmt.Println("Num dobles:", len(listaDoblesPalabrasKeyword))
	//fmt.Println(listaPalabrasKeyword) 
	//return
	// start, err := time.Parse(time.RubyDate, "Thu Apr 24 10:13:37 +0000 2014")
	// end, err := time.Parse(time.RubyDate,"Thu Apr 24 10:15:28 +0000 2014")
	// Primera parte del partido
	start, err := time.Parse(time.RubyDate, "Tue Apr 29 20:45:00 +0200 2014")
	end, err := time.Parse(time.RubyDate,"Tue Apr 29 21:45:00 +0200 2014")
	tweetsFiltrados = trendings.GetTweetsRange(*a, start, end)
	fmt.Println("=============================")
	fmt.Println("Twets filtrados por")
	fmt.Println("=============================")
	fmt.Println("Inicio", start.Unix())
	fmt.Println("Fin", end.Unix())
	fmt.Println("=============================")
	fmt.Println("Num Tweets:", len(tweetsFiltrados))
	// for i:=0; i<len(tweetsFiltrados); i++ {
		// fmt.Println(i, tweetsFiltrados[i].Epoch, tweetsFiltrados[i].Created_at)
	// }
	listaPalabrasRango1, listaDoblesPalabrasRango1 = trendings.ParseTweetsAll(tweetsFiltrados, "#championstve")
	// Segunda parte del partido
	start, err = time.Parse(time.RubyDate, "Tue Apr 29 21:45:01 +0200 2014")
	end, err = time.Parse(time.RubyDate,"Tue Apr 29 22:35:00 +0200 2014")
	tweetsFiltrados = trendings.GetTweetsRange(*a, start, end)
	fmt.Println("=============================")
	fmt.Println("Twets filtrados por")
	fmt.Println("=============================")
	fmt.Println("Inicio", start/*.Unix()*/)
	fmt.Println("Fin", end/*.Unix()*/)
	fmt.Println("=============================")
	fmt.Println("Num Tweets:", len(tweetsFiltrados))
	// for i:=0; i<len(tweetsFiltrados); i++ {
		// fmt.Println(i, tweetsFiltrados[i].Epoch, tweetsFiltrados[i].Created_at)
	// }
	listaPalabrasRango2, listaDoblesPalabrasRango2 = trendings.ParseTweetsAll(tweetsFiltrados, "#championstve")
	
	mergedList = trendings.MergeTrendings(listaPalabras, listaDoblesPalabras)
	// Guardamos los pares [palabra, valor] en la BD
	g := session.DB("twitter").C("resultado1")
	for i:=0; i< len(listaPalabras); i++ {
		//fmt.Println(listaPalabras[i].Value, listaPalabras[i].Word)
		err = g.Insert(listaPalabras[i]) 
		if err != nil { 
			panic(err) 
		}
	}
	
	// for i:=0; i<len(todosLosTweets); i++ {
		// fmt.Println(todosLosTweets[i].Text)
	// }
}