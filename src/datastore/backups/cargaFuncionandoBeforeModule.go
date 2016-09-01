package main

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
	//"profiles"
	"trendings"
	"labix.org/v2/mgo"	
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	//"time"
	"sort"
)

const DatastoreKindName = "surveyresult"

const host = "quetwitteroeres.appspot.com"
const email = "profilizeme@gmail.com"
const password = "profilizeme8230"

var  Intereses = [15]string	{	"famosos", "futbol", "tenis", "baloncesto", "otrosdeportes",
								"tecnologia", "lectura", "cine", "arte", "anime",
								"videojuegos", "musica", "series", "marketing", "moda"			}
var Personalidades = [10]string {	"introvertido", "extrovertido", "agresivo", "calmado", "positivo", 
									"negativo", "lider", "sedejallevar", "metodico", "desordenado" 		}

func main() {
	resultados := GetDatastoreResults()
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	


	// Leemos las timelines y calculamos los mapas de palabras
	interestsWords, personalityWords := TimelinesToMaps(*session, Intereses, Personalidades, resultados)
	
	// Calculamos Lista ordenada de intereses
	interesesOrdenado := MapToListInterests(interestsWords)
	
	// Calculamos el topX de palabras de cada interes
	topWords1 := GetTopInterestsWords (interesesOrdenado, 10)
	
	// Calculamos Lista ordenada de categorias de personalidades
	personalidadesOrdenado := MapToListCategories(personalityWords)
	
	// Calculamos el topX de palabras de cada categoria de personalidad
	topWords2 := GetTopCategoriesWords (personalidadesOrdenado, 10)	
	
	fmt.Println(len(topWords1)+len(topWords2))
	// Guardamos los intereses en la DB 
	//InterestsToDB (interesesOrdenado, Intereses, *session)
	
	// Leemos los intereses de la DB
	interesesOrdenado2 := GetInterestsFromDB (Intereses, *session)
	
	// Borramos los intereses de la DB
	//DropInterests (*session, Intereses)
	
	// Guardamos las categorias de personalidad en DB
	//CategoriesToDB (personalidadesOrdenado, Personalidades, *session)
	
	// Leemos las categorias de personalidad de DB
	personalidadesOrdenado2 := GetCategoriesFromDB (Personalidades, *session)
	
	// Borramos las categorias de personalidad de DB
	//DropCategories (*session, Personalidades)
	
	fmt.Println("---------------------")
	fmt.Println("Datos sacados de bd")
	fmt.Println("---------------------")
	for i:=0; i<len(interesesOrdenado2); i++ {
		fmt.Println(Intereses[i], len(interesesOrdenado2[i]))
	}
	for i:=0; i<len(personalidadesOrdenado2); i++ {
		fmt.Println(Personalidades[i], len(personalidadesOrdenado2[i]))
	}
	
	// Convertimos la lista de intereses sacada de BD a Map
	interestsMap := ListsToMapsInterests (interesesOrdenado2)

	// Convertimos la lista de categorias sacada de BD a map
	categoriesMap := ListsToMapsCategories (personalidadesOrdenado2)
	
	fmt.Println("---------------------")
	fmt.Println("Datos del mapa")
	fmt.Println("---------------------")
	for i:=0; i<len(interestsMap); i++ {
		fmt.Println(Intereses[i], len(interestsMap[i]))
	}
	for i:=0; i<len(categoriesMap); i++ {
		fmt.Println(Personalidades[i], len(categoriesMap[i]))
	}	
}

func MapToListInterests (interestsMap []map[string]int) [15]trendings.WordsList{
	var interOrdenado [15]trendings.WordsList
	for i:=0; i<len(interestsMap); i++ {
		interOrdenado[i] = sortMapByValue(interestsMap[i])
	}
	return interOrdenado
}

func MapToListCategories (categoriesMap []map[string]int) [10]trendings.WordsList{
	var categOrdenado [10]trendings.WordsList
	for i:=0; i<len(categoriesMap); i++ {
		categOrdenado[i] = sortMapByValue(categoriesMap[i])
	}
	return categOrdenado
}

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

func ListsToMapsCategories (lists [10]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{},{},{},{},{},{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

func ListsToMapsInterests (lists [15]trendings.WordsList) []map[string]int{
	var maps = []map[string]int{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}
	for i:=0; i<len(lists); i++ {
		for j:=0; j<len(lists[i]); j++ {
			maps[i][lists[i][j].Word] = lists[i][j].Value
		}
	}
	return maps
}

func CategoriesToDB (categoriesLists [10]trendings.WordsList, personalidades [10]string, session mgo.Session) {
	for i:=0; i<len(personalidades); i++ {
		res1 := session.DB("personalidades").C(personalidades[i])
		for j:=0; j<len(categoriesLists[i]); j++ {
			err := res1.Insert(categoriesLists[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

func InterestsToDB (interestsLists [15]trendings.WordsList, intereses [15]string, session mgo.Session) {
	for i:=0; i<len(intereses); i++ {
		res1 := session.DB("intereses").C(intereses[i])
		for j:=0; j<len(interestsLists[i]); j++ {
			err := res1.Insert(interestsLists[i][j]) 
			if err != nil { 
				panic(err) 
			} 
		}
	}
}	

func GetCategoriesFromDB (personalidades [10]string, session mgo.Session) [10]trendings.WordsList {
	var categoriesLists [10]trendings.WordsList
	for i:=0; i<len(personalidades); i++ {
		res1 := session.DB("personalidades").C(personalidades[i])		
		err := res1.Find(nil).All(&categoriesLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return categoriesLists
}

func GetInterestsFromDB (intereses [15]string, session mgo.Session) [15]trendings.WordsList {
	var interestsLists [15]trendings.WordsList
	for i:=0; i<len(intereses); i++ {
		res1 := session.DB("intereses").C(intereses[i])		
		err := res1.Find(nil).All(&interestsLists[i])
		if err != nil { 
			panic(err) 
		}
	}
	return interestsLists
}

func DropCategories (session mgo.Session, personalidades [10]string) {
	//Deleting the data stored in the DB
	for i:=0; i<len(personalidades); i++ {
		res1 := session.DB("personalidades").C(personalidades[i])
		_ = res1.DropCollection()
	}
}

func DropInterests (session mgo.Session, intereses [15]string) {
	//Deleting the data stored in the DB
	for i:=0; i<len(intereses); i++ {
		res1 := session.DB("intereses").C(intereses[i])
		_ = res1.DropCollection()	
	}
}

func AddWordsToInterests (interestsMap []map[string]int, words trendings.WordsList, interests []string) {
	//fmt.Println(interests)
	for i:=0; i<len(interests); i++ {
		switch interests[i]{
			case "famosos": 			// 0 
				AddWordsToMapGeneric(interestsMap[0], words)
			case "futbol": 				// 1
				AddWordsToMapGeneric(interestsMap[1], words)
			case "tenis":				// 2
				AddWordsToMapGeneric(interestsMap[2], words)
			case "baloncesto":			// 3
				AddWordsToMapGeneric(interestsMap[3], words)
			case "otrosdeportes":		// 4
				AddWordsToMapGeneric(interestsMap[4], words)
			case "tecnologia":			// 5
				AddWordsToMapGeneric(interestsMap[5], words)
			case "lectura":				// 6
				AddWordsToMapGeneric(interestsMap[6], words)
			case "cine":				// 7
				AddWordsToMapGeneric(interestsMap[7], words)
			case "arte":				// 8
				AddWordsToMapGeneric(interestsMap[8], words)
			case "anime":				// 9
				AddWordsToMapGeneric(interestsMap[9], words)
			case "videojuegos":			// 10
				AddWordsToMapGeneric(interestsMap[10], words)
			case "musica":				// 11
				AddWordsToMapGeneric(interestsMap[11], words)
			case "series":				// 12
				AddWordsToMapGeneric(interestsMap[12], words)
			case "marketing":			// 13
				AddWordsToMapGeneric(interestsMap[13], words)
			case "moda":				// 14
				AddWordsToMapGeneric(interestsMap[14], words)
		}
	}
}

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

func AddWordsToPersonality (personalityMap []map[string]int, words trendings.WordsList, timelineScores Scores) {
	// Aqui se podria añadir mas palabras si el score es mayor... de momento si pasa de 6 añadimos las palabras y ya sta
	if timelineScores.Introvertido > 6 		{	AddWordsToMapGeneric(personalityMap[0], words)	}
	if timelineScores.Extrovertido > 6 		{	AddWordsToMapGeneric(personalityMap[1], words)	}
	if timelineScores.Agresivo > 6 			{	AddWordsToMapGeneric(personalityMap[2], words)	}
	if timelineScores.Calmado > 6 			{	AddWordsToMapGeneric(personalityMap[3], words)	}
	if timelineScores.Positivo > 6 			{	AddWordsToMapGeneric(personalityMap[4], words)	}
	if timelineScores.Negativo > 6 			{	AddWordsToMapGeneric(personalityMap[5], words)	}
	if timelineScores.Lider > 6 			{	AddWordsToMapGeneric(personalityMap[6], words)	}
	if timelineScores.Sedejallevar > 6 		{	AddWordsToMapGeneric(personalityMap[7], words)	}
	if timelineScores.Metodico > 6 			{	AddWordsToMapGeneric(personalityMap[8], words)	}
	if timelineScores.Desordenado > 6 		{	AddWordsToMapGeneric(personalityMap[9], words)	}
}
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

func TimelinesToMaps (session mgo.Session, intereses [15]string, personalidades [10]string, resultados []SurveyResult) ([]map[string]int, []map[string]int){
// interestsWords contains a map of [Word, Value] for each interest
	var interestsWords = []map[string]int{{},{},{},{},{},{},{},{},{},{},{},{},{},{},{}}
	// personalityWords contains a map of [Word, Value] for each category
	var personalityWords = []map[string]int{{},{},{},{},{},{},{},{},{},{}}

	//for i:=0; i<len(resultados); i++ { // Version para leer todos los datos
	for i:=len(resultados) - 7; i<len(resultados); i++ { // Version de prueba del bucle
		// Calculamos scores
		timelineScores := CalculateScores(resultados[i])
		// Creamos la coleccion de la BD
		col := session.DB("remoteApi").C(resultados[i].ScreenName)
		fmt.Println("Intentando capturar timeline de: ", resultados[i].ScreenName)
		//profiles.GetTimelineToDB(*col, resultados[i].ScreenName, 0)
		timeline := trendings.GetTweetsAll (*col)
		fmt.Println("Tweets leidos de la bd: ", len(timeline))
	
		timelineWords, timelineDoubleWords := trendings.ParseTweetsAll (timeline, "")
		
		// Calculamos las palabras asociadas a sus intereses
		AddWordsToInterests(interestsWords, timelineWords, resultados[i].Interests)
		AddWordsToInterests(interestsWords, timelineDoubleWords, resultados[i].Interests)

		//fmt.Println(interestsWords[1]["nico abad"])
		// for i:=0; i<len(interestsWords); i++ {
			// fmt.Println("Long del mapa '", intereses[i], ":" , len(interestsWords[i]))
		// }	
		
		// Calculamos las palabras asociadas a sus caracteristicas
		AddWordsToPersonality(personalityWords, timelineWords, timelineScores)
		AddWordsToPersonality(personalityWords, timelineDoubleWords, timelineScores)

		// for i:=0; i<len(personalityWords); i++ {
			// fmt.Println("Long del mapa '", personalidades[i], ":" , len(personalityWords[i]))
		// }			

		//fmt.Println("Long del mapa de Words/Scores: ", len(totalWords))
		// fmt.Println("Palabras/ocurrencias en la timeline: ", len(timelineWords))
		// fmt.Println("Dobles Palabras/ocurrencias en la timeline: ", len(timelineDoubleWords))

		//Deleting the data stored in the DB
		//col.DropCollection()
		// err = col.DropCollection()
		// if err != nil { 
			// panic(err) 
		// }
	}
	return interestsWords, personalityWords
}