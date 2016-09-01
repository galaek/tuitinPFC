package timelines

import (
	"errors"
	//"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	//"strings"
	"appengine"
	"appengine/datastore"
	"appengine/remote_api"
	"fmt"
	"profiles"
	"trendings"
	"labix.org/v2/mgo"	
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"time"
	"strings"
	"strconv"
	"sort"
)

//const DatastoreKindName = "surveyresult"
// DBs funcionando: intereses3, personalidades3, edades3, genero3 
const host = "quetwitteroeres.appspot.com"
const email = "profilizeme@gmail.com"
const password = "profilizeme8230"
const InteresesDB = "intereses3"
const PersonalidadesDB = "personalidades3"
const EdadesDB = "edades3"
const GenerosDB = "genero3"
// -------------------------------------------------
// -------------------------------------------------
// const InteresesDB = "intereses4"
// const PersonalidadesDB = "personalidades4"
// const EdadesDB = "edades4"
// const GenerosDB = "genero4"

var  Intereses = [15]string	{	"famosos", "futbol", "tenis", "baloncesto", "otrosdeportes",
								"tecnologia", "lectura", "cine", "arte", "anime",
								"videojuegos", "musica", "series", "marketing", "moda"			}
var Personalidades = [10]string {	"introvertido", "extrovertido", "agresivo", "calmado", "positivo", 
									"negativo", "lider", "sedejallevar", "metodico", "desordenado" 		}

var Edades = [3]string { 	"under1980", "from1980to1990", "over1990" }
var Generos = [2]string {	"hombre", "mujer"	}

// StringToDate converts a string in format yyyy/mm/dd to time.Time		
func StringToDate (stringedDate string) time.Time {
		fecha := strings.Split(stringedDate, "/")
		year, _ := strconv.ParseInt(fecha[0], 10, 32)
		month, _ := strconv.ParseInt(fecha[1], 10, 32)
		day, _ := strconv.ParseInt(fecha[2], 10, 32)
		return time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC)
}

// DropGender removes from DB the 'genders' lists
func DropGender (session mgo.Session) {
	//Deleting the data stored in the DB
	for i:=0; i<len(Generos); i++ {
		res1 := session.DB(GenerosDB).C(Generos[i])
		_ = res1.DropCollection()	
	}
}

// AddWordsToGender adds the words from the list 'words' to the corresponding map of 'gender'
func AddWordsToGender (genderMap []map[string]int, words trendings.WordsList, gender string) {
	// We calculate totalWords
	totalWords := 0
	for i:=0; i<len(words); i++ {
		totalWords = totalWords + words[i].Value
	}
	normalizedWords := make(trendings.WordsList, len(words)) 
	for i:=0; i<len(words); i++ {
		normalizedWords[i].Word = words[i].Word
		normalizedWords[i].Value = (words[i].Value * 1000000) / totalWords
	}

	if gender == "Hombre" {
		AddWordsToMapGeneric(genderMap[0], normalizedWords)
	} else if gender == "Mujer" {
		AddWordsToMapGeneric(genderMap[1], normalizedWords)
	} else {
		fmt.Println("Someone did something strange about the gender")
	}
	return
}

// MapToListGender Converts the mapss with the words of genders into ordered lists of words
func MapToListGender (genderMap []map[string]int) [2]trendings.WordsList{
	var genderOrdenado [2]trendings.WordsList
	for i:=0; i<len(genderMap); i++ {
		genderOrdenado[i] = sortMapByValue(genderMap[i])
	}
	return genderOrdenado
}

// GenderToDB saves the gender lists into the mongo DB specified by 'session'
func GenderToDB (genderList [2]trendings.WordsList, session mgo.Session) {
	for i:=0; i<len(Generos); i++ {
		res1 := session.DB(GenerosDB).C(Generos[i])
		for j:=0; j<len(genderList[i]); j++ {
			err := res1.Insert(genderList[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

// GetGenderFromDB reads from DB and returns the 'gender' lists
func GetGenderFromDB (session mgo.Session) [2]trendings.WordsList {
	var genderLists [2]trendings.WordsList
	for i:=0; i<len(Generos); i++ {
		res1 := session.DB(GenerosDB).C(Generos[i])		
		err := res1.Find(nil).All(&genderLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return genderLists
}

func ListsToMapsGender (lists [2]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

func CalculateGender (session mgo.Session, resultados []SurveyResult) {
	// genderWords contains a map of [Word, Value] for each gender (hombre, mujer)
	var genderWords = []map[string]int{{},{}}
	var namesMap = map[string]SurveyResult{}
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
	}
	names, _ := session.DB("analizador").CollectionNames()
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		col := session.DB("analizador").C(names[i])
		timeline := trendings.GetTweetsAll (*col)
		timelineWords, _ := trendings.ParseTweetsAll (timeline, "")
		AddWordsToGender(genderWords, timelineWords, namesMap[names[i]].Gender)
	}
	genderOrdenado := MapToListGender(genderWords)
	GenderToDB (genderOrdenado, session)
	return
}


// DropAges removes from DB the 'ages' lists
func DropAges (session mgo.Session) {
	//Deleting the data stored in the DB
	for i:=0; i<len(Edades); i++ {
		res1 := session.DB(EdadesDB).C(Edades[i])
		_ = res1.DropCollection()	
	}
}

// MapToListInterests Converts the maps with the words of interests into ordered lists of words
func MapToListAges (agesMap []map[string]int) [3]trendings.WordsList{
	var agesOrdenado [3]trendings.WordsList
	for i:=0; i<len(agesMap); i++ {
		agesOrdenado[i] = sortMapByValue(agesMap[i])
	}
	return agesOrdenado
}

func ListsToMapsAges (lists [3]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

// AgesToDB saves the ages lists into the mongo DB specified by 'session'
func AgesToDB (agesList [3]trendings.WordsList, session mgo.Session) {
	for i:=0; i<len(Edades); i++ {
		res1 := session.DB(EdadesDB).C(Edades[i])
		for j:=0; j<len(agesList[i]); j++ {
			err := res1.Insert(agesList[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

// AddWordsToAges adds the words from the list 'words' to the corresponding map of 'ages'
func AddWordsToAges (agesMap []map[string]int, words trendings.WordsList, born string) {
	//fmt.Println(interests)
	
	mod := 1
	const shortForm = "2006-Jan-02"
	t1, _ := time.Parse(shortForm, "1980-Jan-01")
	t2, _ := time.Parse(shortForm, "1990-Jan-01")
	date := StringToDate(born)
	if date.Before(t1) {
		mod = 12
	} else if date.Before(t2) {
		mod = 5
	} else {
		mod = 3
	}
	// We calculate totalWords
	totalWords := 0
	for i:=0; i<len(words); i++ {
		totalWords = totalWords + words[i].Value
	}
	normalizedWords := make(trendings.WordsList, len(words)) 
	for i:=0; i<len(words); i++ {
		normalizedWords[i].Word = words[i].Word
		normalizedWords[i].Value = (words[i].Value * 1000000 * mod) / totalWords
	}

	if date.Before(t1) {
		AddWordsToMapGeneric(agesMap[0], normalizedWords)
	} else if date.Before(t2) {
		AddWordsToMapGeneric(agesMap[1], normalizedWords)
	} else {
		AddWordsToMapGeneric(agesMap[2], normalizedWords)
	}
	return
}

// GetAgesFromDB reads from DB and returns the 'ages' lists
func GetAgesFromDB (session mgo.Session) [3]trendings.WordsList {
	var agesLists [3]trendings.WordsList
	for i:=0; i<len(Edades); i++ {
		res1 := session.DB(EdadesDB).C(Edades[i])		
		err := res1.Find(nil).All(&agesLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return agesLists
}
								
func CalculateAges (session mgo.Session, resultados []SurveyResult) {
	// agesWords contains a map of [Word, Value] for each age range (0-1980, 1980-1990, 1990-now)
	var agesWords = []map[string]int{{},{},{}}
	var namesMap = map[string]SurveyResult{}
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
	}
	names, _ := session.DB("analizador").CollectionNames()
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		col := session.DB("analizador").C(names[i])
		timeline := trendings.GetTweetsAll (*col)
		timelineWords, _ := trendings.ParseTweetsAll (timeline, "")
		AddWordsToAges(agesWords, timelineWords, namesMap[names[i]].Born)
	}
	agesOrdenado := MapToListAges(agesWords)
	AgesToDB (agesOrdenado, session)
	return
}
									
// MapToListInterests Converts the maps with the words of interests into ordered lists of words
func MapToListInterests (interestsMap []map[string]int) [15]trendings.WordsList{
	var interOrdenado [15]trendings.WordsList
	for i:=0; i<len(interestsMap); i++ {
		interOrdenado[i] = sortMapByValue(interestsMap[i])
	}
	return interOrdenado
}

// MapToListCategories Converts the maps with the words of categories into ordered lists of words
func MapToListCategories (categoriesMap []map[string]int) [10]trendings.WordsList{
	var categOrdenado [10]trendings.WordsList
	for i:=0; i<len(categoriesMap); i++ {
		categOrdenado[i] = sortMapByValue(categoriesMap[i])
	}
	return categOrdenado
}

// GetDatastoreResults Gets the survey results from the 'datastore' via 'remote_api'
func GetDatastoreResults () []SurveyResult {
	client := clientLoginClient(host, email, password)
	c, err := remote_api.NewRemoteContext(host, client)
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	log.Printf("App ID %q", appengine.AppID(c))
	
	// Connected to 'datastore' via remote_api
	var resultados []SurveyResult
	q := datastore.NewQuery("surveyresult").Limit(50000)

	if _, err := q.GetAll(c, &resultados); err != nil {
		log.Fatalf("Failed to fetch 'surveyresult' info: %v", err)
	}
	return resultados
}

// GetTopCategoriesWords gets the topMax words from the categories lists
func GetTopCategoriesWords (lists [10]trendings.WordsList, max int) [10]trendings.WordsList {
	var top [10]trendings.WordsList 
	max=0
	for i:=0; i<len(lists); i++ {
		if len(lists[i]) > 10 { max = 10 } else { max = len(lists[i]) }
		temp := make(trendings.WordsList, max)
		for j:=0; j<max; j++ {
			temp[j] = lists[i][j]
		}
		top[i] = temp
	}
	return top
}

// GetTopInterestsWords gets the topMax words from the interests lists
func GetTopInterestsWords (lists [15]trendings.WordsList, max int) [15]trendings.WordsList {
	var top [15]trendings.WordsList 
	max=0
	for i:=0; i<len(lists); i++ {
		if len(lists[i]) > 10 { max = 10 } else { max = len(lists[i]) }
		temp := make(trendings.WordsList, max)
		for j:=0; j<max; j++ {
			temp[j] = lists[i][j]
		}
		top[i] = temp
	}
	return top
}

// ListsToMapsCategories converts the lists of categories into maps
func ListsToMapsCategories (lists [10]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{},{},{},{},{},{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

// ListsToMapsInterests converts the lists of interests into maps
func ListsToMapsInterests (lists [15]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

// CategoriesToDB saves the categories lists into the mongo DB specified by 'session'
func CategoriesToDB (categoriesLists [10]trendings.WordsList, session mgo.Session) {
	for i:=0; i<len(Personalidades); i++ {
		res1 := session.DB(PersonalidadesDB).C(Personalidades[i])
		for j:=0; j<len(categoriesLists[i]); j++ {
			err := res1.Insert(categoriesLists[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

// InterestsToDB saves the interests lists into the mongo DB specified by 'session'
func InterestsToDB (interestsLists [15]trendings.WordsList, session mgo.Session) {
	for i:=0; i<len(Intereses); i++ {
		res1 := session.DB(InteresesDB).C(Intereses[i])
		for j:=0; j<len(interestsLists[i]); j++ {
			err := res1.Insert(interestsLists[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

// GetCategoriesFromDB reads from DB and returns the 'categories' lists
func GetCategoriesFromDB (session mgo.Session) [10]trendings.WordsList {
	var categoriesLists [10]trendings.WordsList
	for i:=0; i<len(Personalidades); i++ {
		res1 := session.DB(PersonalidadesDB).C(Personalidades[i])		
		err := res1.Find(nil).All(&categoriesLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return categoriesLists
}

// GetInterestsFromDB reads from DB and returns the 'interests' lists
func GetInterestsFromDB (session mgo.Session) [15]trendings.WordsList {
	var interestsLists [15]trendings.WordsList
	for i:=0; i<len(Intereses); i++ {
		res1 := session.DB(InteresesDB).C(Intereses[i])		
		err := res1.Find(nil).All(&interestsLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return interestsLists
}

// DropCategories removes from DB the 'categories' lists
func DropCategories (session mgo.Session) {
	//Deleting the data stored in the DB
	for i:=0; i<len(Personalidades); i++ {
		res1 := session.DB(PersonalidadesDB).C(Personalidades[i])
		_ = res1.DropCollection()
	}
}

// DropInterests removes from DB the 'interests' lists
func DropInterests (session mgo.Session) {
	//Deleting the data stored in the DB
	for i:=0; i<len(Intereses); i++ {
		res1 := session.DB(InteresesDB).C(Intereses[i])
		_ = res1.DropCollection()	
	}
}

// AddWordsToInterests adds the words from the list 'words' to the corresponding map of interests
func AddWordsToInterests (interestsMap []map[string]int, words trendings.WordsList, interests []string) {
	//fmt.Println(interests)
	// We calculate totalWords
	totalWords := 0
	for i:=0; i<len(words); i++ {
		totalWords = totalWords + words[i].Value
	}
	fmt.Println("Palabras totales:", totalWords)
	normalizedWords := make(trendings.WordsList, len(words)) 
	mod := 1
	for i:=0; i<len(words); i++ {
		mod = 1
		if i == 0 {mod = 3}
		if i == 1 {mod = 3}
		if i == 2 {mod = 12}
		if i == 3 {mod = 6}
		if i == 4 {mod = 4}
		if i == 5 {mod = 1}
		if i == 6 {mod = 3}
		if i == 7 {mod = 3}
		if i == 8 {mod = 6}
		if i == 9 {mod = 24}
		if i == 10 {mod = 4}
		if i == 11 {mod = 2}
		if i == 12 {mod = 2}
		if i == 13 {mod = 6}
		if i == 14 {mod = 3}
		normalizedWords[i].Word = words[i].Word
		normalizedWords[i].Value = (words[i].Value * 1000000 * mod) / totalWords
	}
	for i:=0; i<len(interests); i++ {

		switch interests[i]{
			case "famosos": 			// 0 
				AddWordsToMapGeneric(interestsMap[0], normalizedWords)
			case "futbol": 				// 1
				AddWordsToMapGeneric(interestsMap[1], normalizedWords)
			case "tenis":				// 2
				AddWordsToMapGeneric(interestsMap[2], normalizedWords)
			case "baloncesto":			// 3
				AddWordsToMapGeneric(interestsMap[3], normalizedWords)
			case "otrosdeportes":		// 4
				AddWordsToMapGeneric(interestsMap[4], normalizedWords)
			case "tecnologia":			// 5
				AddWordsToMapGeneric(interestsMap[5], normalizedWords)
			case "lectura":				// 6
				AddWordsToMapGeneric(interestsMap[6], normalizedWords)
			case "cine":				// 7
				AddWordsToMapGeneric(interestsMap[7], normalizedWords)
			case "arte":				// 8
				AddWordsToMapGeneric(interestsMap[8], normalizedWords)
			case "anime":				// 9
				AddWordsToMapGeneric(interestsMap[9], normalizedWords)
			case "videojuegos":			// 10
				AddWordsToMapGeneric(interestsMap[10], normalizedWords)
			case "musica":				// 11
				AddWordsToMapGeneric(interestsMap[11], normalizedWords)
			case "series":				// 12
				AddWordsToMapGeneric(interestsMap[12], normalizedWords)
			case "marketing":			// 13
				AddWordsToMapGeneric(interestsMap[13], normalizedWords)
			case "moda":				// 14
				AddWordsToMapGeneric(interestsMap[14], normalizedWords)
		}
	}
}

// AddWordsToMapGeneric adds the words from 'words' to the map 'm'
func AddWordsToMapGeneric(m map[string]int, words trendings.WordsList) {
	for i:=0; i<len(words); i++ {
		value, exist := m[words[i].Word]
		if exist == false {
			m[words[i].Word] = words[i].Value
		} else {
			m[words[i].Word] = words[i].Value + value
		}
	}
}

// AddWordsToPersonality adds the words from the list 'words' to the corresponding map of 'categories'
func AddWordsToPersonality (personalityMap []map[string]int, words trendings.WordsList, timelineScores Scores) {
	// We calculate totalWords
	totalWords := 0
	for i:=0; i<len(words); i++ {
		totalWords = totalWords + words[i].Value
	}
	mod := 1
	fmt.Println("Palabras totales:", totalWords)
	normalizedWords := make(trendings.WordsList, len(words)) 
	for i:=0; i<len(words); i++ {
		mod = 1
		if i == 0 {mod = 2}
		if i == 2 {mod = 2}
		if i == 5 {mod = 3}
		if i == 7 {mod = 2}
		if i == 8 {mod = 3}
		
		normalizedWords[i].Word = words[i].Word
		normalizedWords[i].Value = (words[i].Value * 1000000 * mod) / totalWords
	}
	
	
	// Aqui se podria añadir mas palabras si el score es mayor... de momento si pasa de 6 añadimos las palabras y ya sta
	if timelineScores.Introvertido > 6 		{	AddWordsToMapGeneric(personalityMap[0], normalizedWords)	}
	if timelineScores.Extrovertido > 6 		{	AddWordsToMapGeneric(personalityMap[1], normalizedWords)	}
	if timelineScores.Agresivo > 6 			{	AddWordsToMapGeneric(personalityMap[2], normalizedWords)	}
	if timelineScores.Calmado > 6 			{	AddWordsToMapGeneric(personalityMap[3], normalizedWords)	}
	if timelineScores.Positivo > 6 			{	AddWordsToMapGeneric(personalityMap[4], normalizedWords)	}
	if timelineScores.Negativo > 6 			{	AddWordsToMapGeneric(personalityMap[5], normalizedWords)	}
	if timelineScores.Lider > 6 			{	AddWordsToMapGeneric(personalityMap[6], normalizedWords)	}
	if timelineScores.Sedejallevar > 6 		{	AddWordsToMapGeneric(personalityMap[7], normalizedWords)	}
	if timelineScores.Metodico > 6 			{	AddWordsToMapGeneric(personalityMap[8], normalizedWords)	}
	if timelineScores.Desordenado > 6 		{	AddWordsToMapGeneric(personalityMap[9], normalizedWords)	}
}

// CalculateScores uses the survey results to calculate the scores for each 'category'
func CalculateScores (result SurveyResult) Scores {
	var score = Scores {0,0,0,0,0,0,0,0,0,0}
	// Podria multiplicar la puntuacion de las situaciones por 2 o 3 para darles mas peso
	score.Introvertido = 6 - result.Personal[0] + result.Situation[0]
	score.Extrovertido = result.Personal[0] + 6 - result.Situation[0]
	score.Agresivo = 6 - result.Personal[1] + 6 - result.Situation[1]
	score.Calmado = result.Personal[1] + result.Situation[1]
	score.Positivo = result.Personal[2] + 6 - result.Situation[2]
	score.Negativo = 6 - result.Personal[2] + result.Situation[2]
	score.Lider = result.Personal[3] + 6 - result.Situation[3]
	score.Sedejallevar = 6 - result.Personal[3] + result.Situation[3]
	score.Metodico = result.Personal[4] + 6 - result.Situation[4]
	score.Desordenado = 6 - result.Personal[4] + result.Situation[4]
	return score
}

type Scores struct {
	Introvertido	int
	Extrovertido	int
	Agresivo		int
	Calmado			int
	Positivo		int
	Negativo		int
	Lider			int	
	Sedejallevar	int
	Metodico		int
	Desordenado		int
	//Ocurrencias		int
}

type TwitterInfo struct {
	ScreenName		string
	AccessKey		string
	SecretKey		string
	UserID			int64
}

type ProfileInfo struct {
	Nombre		string
	Descripción	string
	Hashtag		string
	ImagenURL	string
	ScreenName	string
	NombrePlano	string
}

type SurveyResult struct {
	ScreenName	string
	Born		string
	Country		string
	Gender		string
	Personal	[]int
	Situation	[]int
	Interests	[]string
	Scores		[]int
	Profile		int
}

// CheckDuplicates returns the list of duplicated ScreenNames in the 'surveyresult' pool
// NOTE: After the update of the survey, there should not be any duplicate entry
func CheckDuplicates (data []SurveyResult) []string {
	var duplicates []string
	var all []string
	for i:=0; i<len(data); i++ {
		if AlreadyExists(all, data[i].ScreenName) {
			duplicates = append(duplicates, data[i].ScreenName)
		} else {
			all = append(all, data[i].ScreenName)
		}
	}
	return duplicates
}

// AlreadyExists returns true if 'word' is in the list 'data', returns false otherwise
func AlreadyExists (data []string, word string) bool {
	for i:=0; i<len(data); i++ {
		if data[i] == word {
			return true
		}
	}
	return false
}

// sortMapByValue converts the map into a sorted list
func sortMapByValue(m map[string]int) trendings.WordsList {
   p := make(trendings.WordsList, len(m))
   i := 0
   for k, v := range m {
      p[i] = trendings.Pair{k, v}
	  i++
   }
   sort.Sort(p)
   return p
}

// clientLoginClient connects to the remote_api to get data from 'datastore'
func clientLoginClient(host, email, password string) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("failed to make cookie jar: %v", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	v := url.Values{}
	v.Set("Email", email)
	v.Set("Passwd", password)
	v.Set("service", "ah")
	//v.Set("source", "Misc-remote_api-0.1")
	v.Set("source", "Google-remote_api-1.0")
	v.Set("accountType", "HOSTED_OR_GOOGLE")

	resp, err := client.PostForm("https://www.google.com/accounts/ClientLogin", v)
	if err != nil {
		log.Fatalf("could not post login: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unsuccessful request: status %d; body %q", resp.StatusCode, body)
	}
	if err != nil {
		log.Fatalf("unable to read response: %v", err)
	}

	m := regexp.MustCompile(`Auth=(\S+)`).FindSubmatch(body)
	if m == nil {
		log.Fatalf("no auth code in response %q", body)
	}
	auth := string(m[1])

	u := &url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     "/_ah/login",
		RawQuery: "continue=/&auth=" + url.QueryEscape(auth),
	}

	// Disallow redirects.
	redirectErr := errors.New("stopping redirect")
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectErr
	}

	resp, err = client.Get(u.String())
	if urlErr, ok := err.(*url.Error); !ok || urlErr.Err != redirectErr {
		log.Fatalf("could not get auth cookies: %v", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusFound {
		log.Fatalf("unsuccessful request: status %d; body %q", resp.StatusCode, body)
	}

	client.CheckRedirect = nil
	return client
}

func GetAllTimelinesToDB (session mgo.Session, resultados []SurveyResult) {
	names, _ := session.DB("analizador").CollectionNames()
	// We clean the existing timelines
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		col := session.DB("analizador").C(names[i])
		col.DropCollection()
	}
	// We get the new timelines
	for i:=0; i<len(resultados); i++ {
		col := session.DB("analizador").C(resultados[i].ScreenName)
		profiles.GetTimelineToDB(*col, resultados[i].ScreenName, 0)
	}
}

func CalculateInterestsAndCategories (session mgo.Session, resultados []SurveyResult) {
	// interestsWords contains a map of [Word, Value] for each interest
	var interestsWords = []map[string]int{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}
	// personalityWords contains a map of [Word, Value] for each category
	var personalityWords = []map[string]int{{},{},{},{},{},{},{},{},{},{}}
	var namesMap = map[string]SurveyResult{}
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
	}
	names, _ := session.DB("analizador").CollectionNames()
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		col := session.DB("analizador").C(names[i])
		timeline := trendings.GetTweetsAll (*col)
		timelineWords, _ := trendings.ParseTweetsAll (timeline, "")
		AddWordsToInterests(interestsWords, timelineWords, namesMap[names[i]].Interests)
		timelineScores := CalculateScores(namesMap[names[i]])
		AddWordsToPersonality(personalityWords, timelineWords, timelineScores)
	}
	interesesOrdenado := MapToListInterests(interestsWords)
	personalidadesOrdenado := MapToListCategories(personalityWords)
	InterestsToDB (interesesOrdenado, session)
	CategoriesToDB (personalidadesOrdenado, session)
}

// TimelinesToMaps **DEPRECATED** reads the timelines from Twitter and stores them in DB. 
// Then calculates the interests and categories words and doublewords
func TimelinesToMaps (session mgo.Session, resultados []SurveyResult) ([]map[string]int, []map[string]int){
	// interestsWords contains a map of [Word, Value] for each interest
	var interestsWords = []map[string]int{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}
	// personalityWords contains a map of [Word, Value] for each category
	var personalityWords = []map[string]int{{},{},{},{},{},{},{},{},{},{}}

	for i:=0; i<len(resultados); i++ { // Version para leer todos los datos
	//for i:=len(resultados) - 7; i<len(resultados); i++ { // Version de prueba del bucle
		// Calculamos scores
		timelineScores := CalculateScores(resultados[i])
		// Creamos la coleccion de la BD
		col := session.DB("remoteApi").C(resultados[i].ScreenName)
		fmt.Println("Intentando capturar timeline de: ", resultados[i].ScreenName)
		profiles.GetTimelineToDB(*col, resultados[i].ScreenName, 0)
		timeline := trendings.GetTweetsAll (*col)
		fmt.Println("Tweets leidos de la bd: ", len(timeline))
		timelineWords, _ := trendings.ParseTweetsAll (timeline, "")
		
		// Calculamos las palabras asociadas a sus intereses
		AddWordsToInterests(interestsWords, timelineWords, resultados[i].Interests)
		//AddWordsToInterests(interestsWords, timelineDoubleWords, resultados[i].Interests)

		// Calculamos las palabras asociadas a sus caracteristicas
		AddWordsToPersonality(personalityWords, timelineWords, timelineScores)
		//AddWordsToPersonality(personalityWords, timelineDoubleWords, timelineScores)
		
		//Deleting the data stored in the DB
		col.DropCollection()
	}
	return interestsWords, personalityWords
}


// AnalyzeTimeline gets a ScreenName and analyzes the Twitter Profile 
func AnalyzeExistingTimeline (screenName string, interestsMaps []map[string]int, categoriesMaps []map[string]int, agesMaps []map[string]int, genderMaps []map[string]int, session mgo.Session) ResultScore {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra con todos los intereses y categorias y sumar la puntuacion donde corresponda
	var res ResultScore
	c := session.DB("analizador").C(screenName)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")

	for j:=0; j<15; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := interestsMaps[j][words[i].Word]
			if exist == true {
				res.Interests[j] = res.Interests[j] + words[i].Value * value
			}
		}
	}
	for j:=0; j<10; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := categoriesMaps[j][words[i].Word]
			if exist == true {
				res.Categories[j] = res.Categories[j] + words[i].Value * value
			}
		}
	}
	for j:=0; j<3; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := agesMaps[j][words[i].Word]
			if exist == true {
				res.Ages[j] = res.Ages[j] + words[i].Value * value
			}
		}
	}
	for j:=0; j<2; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := genderMaps[j][words[i].Word]
			if exist == true {
				res.Gender[j] = res.Gender[j] + words[i].Value * value
			}
		}
	}
	return res
}

// AnalyzeTimeline gets a ScreenName and analyzes the Twitter Profile 
func AnalyzeTimeline (screenName string, interestsMaps []map[string]int, categoriesMaps []map[string]int, agesMaps []map[string]int, genderMaps []map[string]int, session mgo.Session) ResultScore {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra con todos los intereses y categorias y sumar la puntuacion donde corresponda
	var res ResultScore
	c := session.DB("analisis").C(screenName)
	c.DropCollection() // In case we already saved it before
	profiles.GetTimelineToDB(*c, screenName, 0)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")

	for j:=0; j<15; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := interestsMaps[j][words[i].Word]
			if exist == true {
				res.Interests[j] = res.Interests[j] + words[i].Value * value
			}
		}
	}
	for j:=0; j<10; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := categoriesMaps[j][words[i].Word]
			if exist == true {
				res.Categories[j] = res.Categories[j] + words[i].Value * value
			}
		}
	}
	for j:=0; j<3; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := agesMaps[j][words[i].Word]
			if exist == true {
				res.Ages[j] = res.Ages[j] + words[i].Value * value
			}
		}
	}
	
	for j:=0; j<2; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := genderMaps[j][words[i].Word]
			if exist == true {
				res.Gender[j] = res.Gender[j] + words[i].Value * value
			}
		}
	}
	return res
}

// AnalyzeGenderTimeline gets a ScreenName timeline and analyzes the gender of its Twitter Profile 
func AnalyzeGenderTimeline (screenName string, genderMaps []map[string]int, session mgo.Session) [2]int {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra y sumar la puntuacion donde corresponda
	var res [2]int
	c := session.DB("analisis").C(screenName)
	c.DropCollection() // In case we already saved it before
	profiles.GetTimelineToDB(*c, screenName, 0)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")
	for j:=0; j<2; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := genderMaps[j][words[i].Word]
			if exist == true {
				res[j] = res[j] + words[i].Value * value
			}
		}
	}
	return res
}

// AnalyzeGenderExistingTimeline gets a ScreenName and analyzes the gender of the Twitter Profile 
func AnalyzeGenderExistingTimeline (screenName string, genderMaps []map[string]int, session mgo.Session) [2]int {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra y sumar la puntuacion en el genero que corresponda
	var res [2]int
	c := session.DB("analizador").C(screenName)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")
	for j:=0; j<2; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := genderMaps[j][words[i].Word]
			if exist == true {
				res[j] = res[j] + words[i].Value * value
			}
		}
	}
	return res
}

// AnalyzeAgeTimeline gets a ScreenName timeline and analyzes the age of its Twitter Profile 
func AnalyzeAgeTimeline (screenName string, agesMaps []map[string]int, session mgo.Session) [3]int {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra y sumar la puntuacion donde corresponda
	var res [3]int
	c := session.DB("analisis").C(screenName)
	c.DropCollection() // In case we already saved it before
	profiles.GetTimelineToDB(*c, screenName, 0)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")
	for j:=0; j<3; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := agesMaps[j][words[i].Word]
			if exist == true {
				res[j] = res[j] + words[i].Value * value
			}
		}
	}
	return res
}

// AnalyzeAgesExistingTimeline gets a ScreenName and analyzes the Twitter Profile 
func AnalyzeAgesExistingTimeline (screenName string, agesMaps []map[string]int, session mgo.Session) [3]int {
	// Pasos a seguir
	// Leer timeline de 'screenName' profiles.GetTimelineToDB
	// Leer todos los tweets y convertirlos a [palabra, ocurrencias] gettweetsall + trendings.ParseTweetsAll
	// Comparar cada palabra con todos los intereses y categorias y sumar la puntuacion donde corresponda
	var res [3]int
	c := session.DB("analizador").C(screenName)
	timeline := trendings.GetTweetsAll (*c)
	words, _ := trendings.ParseTweetsAll(timeline, "")
	for j:=0; j<3; j++ {
		for i:=0; i<len(words); i++ { 
			value, exist := agesMaps[j][words[i].Word]
			if exist == true {
				res[j] = res[j] + words[i].Value * value
			}
		}
	}
	return res
}

type ResultScore struct {
	Interests		[15]int
	Categories		[10]int
	Ages			[3]int
	Gender			[2]int
}

// NormalizeResults returns the results converted to percents
func NormalizeResults (res ResultScore) ResultScore {
	result := ResultScore{}
	tot1 := 0
	for i:=0; i<15; i++ {
		tot1 = tot1 + res.Interests[i]
	} 
	if tot1 == 0 {return res}
	for i:=0; i<15; i++ {
		result.Interests[i] = res.Interests[i]*100 / tot1
	}
	
	tot := 0
	for i:=0; i<10; i = i + 2 {
		tot = res.Categories[i] + res.Categories[i+1]
		result.Categories[i] = res.Categories[i]*100 / tot
		result.Categories[i+1] = res.Categories[i+1]*100 / tot
	}	
	
	tot2 := 0
	for i:=0; i<3; i++ {
		tot2 = tot2 + res.Ages[i]
	} 
	if tot2 == 0 {return res}
	for i:=0; i<3; i++ {
		result.Ages[i] = res.Ages[i]*100 / tot2
	}
	
	tot3 := 0
	for i:=0; i<2; i++ {
		tot3 = tot3 + res.Gender[i]
	} 
	if tot3 == 0 {return res}
	for i:=0; i<2; i++ {
		result.Gender[i] = res.Gender[i]*100 / tot3
	}
	return result
}

// NormalizeAgeResults returns the results converted to percents
func NormalizeAgeResults (res [3]int) [3]int {
	var result [3]int
	//result:=make(int, 3)
	tot1 := 0
	for i:=0; i<3; i++ {
		tot1 = tot1 + res[i]
	} 
	if tot1 == 0 {return res}
	for i:=0; i<3; i++ {
		result[i] = res[i]*100 / tot1
	}
	return result
}

// NormalizeGenderResults returns the results converted to percents
func NormalizeGenderResults (res [2]int) [2]int {
	var result [2]int
	tot1 := 0
	for i:=0; i<2; i++ {
		tot1 = tot1 + res[i]
	} 
	if tot1 == 0 {return res}
	for i:=0; i<2; i++ {
		result[i] = res[i]*100 / tot1
	}
	return result
}

// PrintGenderResults prints the results
func PrintGenderResults (res [2]int) {
	fmt.Println("---------")
	fmt.Println("Genero:")
	fmt.Println("---------")
	for i:=0; i<2; i++ {
		fmt.Println(res[i], "%", "=", Generos[i])
	}

}

// PrintAgesResults prints the results
func PrintAgesResults (res [3]int) {
	fmt.Println("---------")
	fmt.Println("Edad:")
	fmt.Println("---------")
	for i:=0; i<3; i++ {
		fmt.Println(res[i], "%", "=", Edades[i])
	}

}

// PrintResults prints the results
func PrintResults (res ResultScore) {
	fmt.Println("----------")
	fmt.Println("Intereses:")
	fmt.Println("----------")
	for i:=0; i<15; i++ {
		fmt.Println(Intereses[i], "=", res.Interests[i], "%")
	}
	fmt.Println("-------------")
	fmt.Println("Personalidad:")
	fmt.Println("-------------")
	for i:=0; i<10; i++ {
		fmt.Println(Personalidades[i], "=", res.Categories[i], "%")
	}
	fmt.Println("---------")
	fmt.Println("Edad:")
	fmt.Println("---------")
	for i:=0; i<3; i++ {
		fmt.Println(res.Ages[i], "%", "=", Edades[i])
	}
	fmt.Println("---------")
	fmt.Println("Genero:")
	fmt.Println("---------")
	for i:=0; i<2; i++ {
		fmt.Println(res.Gender[i], "%", "=", Generos[i])
	}
}

// PrintOrderedDates prints the list of 'born' dates in datastore. 
// 	This function was used to see the dates ordered and chose the ranges.
func PrintOrderedDates(resultados []SurveyResult) {
	const shortForm = "2006-Jan-02"
	fechas := make([]time.Time, len(resultados))
	for i:=0; i<len(resultados); i++ {
		fecha := strings.Split(resultados[i].Born, "/")
		year, _ := strconv.ParseInt(fecha[0], 10, 32)
		month, _ := strconv.ParseInt(fecha[1], 10, 32)
		day, _ := strconv.ParseInt(fecha[2], 10, 32)
		fechas[i] = time.Date(int(year), time.Month(int(month)), int(day), 0, 0, 0, 0, time.UTC)
	}
	sort.Sort(dateList(fechas))
	fmt.Println("Num fechas en datastore:", len(fechas))

	t1, _ := time.Parse(shortForm, "1980-Jan-01")
	t2, _ := time.Parse(shortForm, "1990-Jan-01")
	for i:=0; i<len(fechas); i++ {
		if fechas[i].Before(t1) {
			fmt.Println("Antes del 80:", fechas[i])
		} else if fechas[i].Before(t2) {
			fmt.Println("Antes del 90:", fechas[i])
		} else {
			fmt.Println("Despues del 90:", fechas[i])
		}
	}
	return	
}

type dateList []time.Time
func (a dateList) Len() int           { return len(a) }
func (a dateList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a dateList) Less(i, j int) bool { return a[i].Before(a[j]) }