package main

// Using application-only auth.

import (
	"flag"
	"fmt"
	"twittertypes"
	"strings"
	"labix.org/v2/mgo"	
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"strconv"
	"profiles"
	"trendings"

)

type Args struct {
	ScreenName string
	OutputFile string
}

func parseArgs() *Args {
	a := &Args{}
	flag.StringVar(&a.ScreenName, "screen_name", "twitterapi", "Screen name")
	flag.StringVar(&a.OutputFile, "out", "timeline.json", "Output file")
	flag.Parse()
	return a
}

func main() {
	var (
		err     error
		args    *Args
	)
	
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	c := session.DB("timelines").C("terminal")
	args = parseArgs()
	

	// Test the double words
	//DoubleWordsTesting()
	// Determine Age Range
	//AgeRangeTesting(*session)
	// Asking Twiter for the timeline
	//profiles.GetTimelineToDB(*c, args.ScreenName, 500) // 2, 9, 25, 30, 40....
	//profiles.GetTimelineToDB(*c, "galaek", 500)
	// Calculating the number of tweets per day
	//TweetsPerDayTesting(*c)
	// Deleting the data stored in the DB
	//err = c.DropCollection()
	//if err != nil { 
	//	panic(err) 
	//}
	//Getting a Twitter User
	//user := UserProfileTesting()
	// Determine Gender
	// GenderTesting(user.Name)
	// GenderTesting("Hipopótamo Pepita")
	// GenderTesting("Hipopótamo Pepita")
	// gen, err := profiles.GenderFromName("María")
	// if err != nil {
		// fmt.Println(err)
	// } else {
		// fmt.Println(gen)
	// }
	// Get Following List
	//GetFriendsTesting(*session)
	// Get Followers List
	//GetFollowersTesting(*session)
	// Get Country and state ("Provincia" for spanish places) from [lat,long]
	//GetCountryStateTesting(*session)
	GetLocationTesting()
	
	return 
	fmt.Println(args)
	fmt.Println(c)
}

// Testing functions

func DoubleWordsTesting () {
	// DoubleWords test
	palabras := trendings.TextToWords("\"Un vestuario sano vale más que cien horas de táctica\" Vicente del Bosque ")
	for i:=0; i<len(palabras); i++ {
		fmt.Printf("-->%s<--\n", palabras[i])
	}
	fmt.Println("-------------")
	palabrasFiltradas := trendings.WordsFilter(palabras, "#shit")
	for i:=0; i<len(palabrasFiltradas); i++ {
		fmt.Printf("-->%s<--\n", palabrasFiltradas[i])
	}	
	fmt.Println("-------------")
	resultado := trendings.WordsToDoubleWords(palabrasFiltradas)
	for i:=0; i<len(resultado); i++ {
		fmt.Printf("-->%s<--\n", resultado[i])
	}
	fmt.Println("-------------")
}

func AgeRangeTesting (session mgo.Session) {
	 
	c := session.DB("timelines").C("edades")
	//profiles.GetTimelineToDB(*c, "srbuhito", 500) 
	profiles.GetTimelineToDB(*c, "galaek", 500)
	todosLosTweets := trendings.GetTweetsAll(*c)
	listaPalabras, _ := trendings.ParseTweetsAll(todosLosTweets, "")
	//fmt.Println(listaPalabras)
	counter1 := 0;
	counter2 := 0;
	counter3 := 0;
	counter4 := 0;
	for i:=0; i<len(listaPalabras); i++ {
		value, exists := profiles.WordsDB_age13to18[listaPalabras[i].Word]
		if exists == true {
			counter1 = counter1 + (listaPalabras[i].Value * value)
		}
		value, exists = profiles.WordsDB_age19to22[listaPalabras[i].Word]
		if exists == true {
			counter2 = counter2 + (listaPalabras[i].Value * value)
		}
		value, exists = profiles.WordsDB_age23to29[listaPalabras[i].Word]
		if exists == true {
			counter3 = counter3 + (listaPalabras[i].Value * value)
		}
		value, exists = profiles.WordsDB_age30to65[listaPalabras[i].Word]
		if exists == true {
			counter4 = counter4 + (listaPalabras[i].Value * value)
		}
	}
	fmt.Println("Puntos de 13 a 18:", counter1)
	fmt.Println("Puntos de 19 a 22:", counter2)
	fmt.Println("Puntos de 23 a 29:", counter3)
	fmt.Println("Puntos de 30 a 65:", counter4)
	err := c.DropCollection()
	if err != nil { 
		panic(err) 
	}	
}

func GetLocationTesting () {
	var user1 twittertypes.UserProfile
	user1 = profiles.GetUserByScreenName("soniapuy")
	//user1 = profiles.GetUserByScreenName("galaek")
	fmt.Println("Location:", profiles.GetLocation(user1))
}

func GetCountryStateTesting (session mgo.Session) {
	place1, err := profiles.GetCountryAndStateByCoordinates(41.512088, -0.629692)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("-------")
		fmt.Println("Home:")
		fmt.Println("Country:", place1.Country)
		fmt.Println("State:", place1.State)
		fmt.Println("-------")
	}
	place1, err = profiles.GetCountryAndStateByCoordinates(40.2531125, -3.706598)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Pinto (Madrid):")
		fmt.Println("Country:", place1.Country)
		fmt.Println("State:", place1.State)
		fmt.Println("-------")
	}
	place1, err = profiles.GetCountryAndStateByCoordinates(50.73394, 4.24208)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Halle, Belgium:")
		fmt.Println("Country:", place1.Country)
		fmt.Println("State:", place1.State)
		fmt.Println("-------")
	}
	c := session.DB("timelines").C("coordinates")
	tweets := trendings.GetTweetsAll(*c)	
	place1, err = profiles.GetCountryAndStateByCoordinates(tweets[0].Coordinates.Coordinates[1], tweets[0].Coordinates.Coordinates[0])
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Tweet From CEEI (Zaragoza):")
		fmt.Println("Country:", place1.Country)
		fmt.Println("State:", place1.State)
		fmt.Println("-------")
	}
}

func GetFollowersTesting(session mgo.Session) {
	var user1 twittertypes.UserProfile
	var screen_name string
	// Get by screen name
	screen_name = "Adrimariscal"
	g:= session.DB("followersof").C(screen_name)
	profiles.GetFollowersByScreenNameToDB(*g, screen_name)
	// Get by id
	screen_name = "luisocro"
	g = session.DB("followersof").C(screen_name)
	user1 = profiles.GetUserByScreenName(screen_name)
	profiles.GetFollowersByIDToDB(*g, user1.Id)	
}

func GetFriendsTesting(session mgo.Session) {
	var user1 twittertypes.UserProfile
	var screen_name string
	// Get by screen name
	screen_name = "SSantiagosegura"
	g:= session.DB("friendsof").C(screen_name)
	profiles.GetFriendsByScreenNameToDB(*g, screen_name)
	// Get by id
	screen_name = "As_TomasRoncero"
	g = session.DB("friendsof").C(screen_name)
	user1 = profiles.GetUserByScreenName(screen_name)
	profiles.GetFriendsByIDToDB(*g, user1.Id)	
}

func GenderTesting (name string) {
	var gen profiles.Gender
	gen, err := profiles.GenderFromName(name)
	if err != nil {
		fmt.Println(err)
	} else {
		if gen.Gender == "null" {
			fmt.Println("Complete name not found")
			names := strings.Split(name, " ")
			for i:=0; i<len(names); i++ {
				gen, err = profiles.GenderFromName(names[i])
				if err != nil {
					fmt.Println(err)
					break
				} else {
					if (gen.Gender != "null") && (gen.Probability > 0.5) {
						fmt.Println("Found match on part:", i)
						break
					}
				}
			}
		} else {
			fmt.Println("Complete name found")
		}
		if gen.Gender == "null" {
			fmt.Println("Unable to determine gender")
		} else {
			if gen.Probability < 0.5 {
				fmt.Println("Gender =", gen.Gender, "(with low probability)")
			} else {
				fmt.Println("Gender =", gen.Gender, "(pretty sure)")
			}
		}	
	}
}

func UserProfileTesting() twittertypes.UserProfile{
	var user1 twittertypes.UserProfile
	user1 = profiles.GetUserByScreenName("As_TomasRoncero")
	//user1 = profiles.GetUserByScreenName("galaek")
	fmt.Println("ScreenName:", user1.Name)
	fmt.Println("StatusText:", user1.Status.Text)
	fmt.Println("StringedID:", strconv.FormatInt(user1.Id, 10))
	var user2 twittertypes.UserProfile
	user2 = profiles.GetUserByID(user1.Id)
	fmt.Println("ScreenName:", user2.Name)
	fmt.Println("StatusText:", user2.Status.Text)
	fmt.Println("Id:", user2.Id)
	return user1
}

func TweetsPerDayTesting(c mgo.Collection) {
	var tweetsPerDay profiles.TweetsFrequency
	tweetsPerDay = profiles.GetTweetsPerDay(c)
	//fmt.Println(tweetsPerDay)
	fmt.Println("Average tweets/day:", tweetsPerDay.Average)
	fmt.Printf("[ ")
	for i:=0; i<len(tweetsPerDay.List); i++ {
		fmt.Printf("%d ", tweetsPerDay.List[i].NumTweets)
	}
	fmt.Printf("] \n")
	fmt.Println("Total Days:", tweetsPerDay.TotalDays)
	fmt.Println("Total Tweets:", tweetsPerDay.TotalTweets)
	return
}
