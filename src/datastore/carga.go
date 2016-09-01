package main

import (
	//"errors"
	//"flag"
	//"io/ioutil"
	//"log"
	//"net/http"
	//"net/http/cookiejar"
	//"net/url"
	//"regexp"
	//"strings"
	//"appengine"
	//"appengine/datastore"
	//"appengine/remote_api"
	"fmt"
	//"profiles"
	//"trendings"
	"timelines"
	"labix.org/v2/mgo"	
	//"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	//"time"
	//"sort"
)

func main() {
	
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	// Por si queremos ver si hay entradas duplicadas en 'datastore'
	// Despues del update ya no deberia de haber duplicados.
	// resultados := timelines.GetDatastoreResults()
	// duplicados := timelines.CheckDuplicates(resultados)
	// fmt.Println("Duplicados:", len(duplicados))
	// fmt.Println("Lista de duplicados:", duplicados)
	// return
	
	/*
	// Obtenemos la info del datastore
	resultados := timelines.GetDatastoreResults()
	
	// Leemos las timelines y calculamos los mapas de palabras
	interestsWords, personalityWords := timelines.TimelinesToMaps(*session, resultados)
	
	// Calculamos Lista ordenada de intereses
	interesesOrdenado := timelines.MapToListInterests(interestsWords)
	// Calculamos Lista ordenada de categorias de personalidades
	personalidadesOrdenado := timelines.MapToListCategories(personalityWords)
	
	// Calculamos el topX de palabras de cada interes
	topWords1 := timelines.GetTopInterestsWords (interesesOrdenado, 10)
	// Calculamos el topX de palabras de cada categoria de personalidad
	topWords2 := timelines.GetTopCategoriesWords (personalidadesOrdenado, 10)	
	
	fmt.Println(len(topWords1)+len(topWords2))
	
	// Guardamos los intereses en la DB 
	timelines.InterestsToDB (interesesOrdenado, *session)
	// Guardamos las categorias de personalidad en DB
	timelines.CategoriesToDB (personalidadesOrdenado, *session)
	
	return
	*/
	
	// Leemos los intereses de la DB
	//interesesOrdenado2 := timelines.GetInterestsFromDB (*session)
	// Leemos las categorias de personalidad de DB
	//personalidadesOrdenado2 := timelines.GetCategoriesFromDB (*session)
	
	/*
	// Calculamos el topX de palabras de cada interes
	topWords1 := timelines.GetTopInterestsWords (interesesOrdenado2, 10)
	// Calculamos el topX de palabras de cada categoria de personalidad
	topWords2 := timelines.GetTopCategoriesWords (personalidadesOrdenado2, 10)
	
	*/
	
	// Borramos los intereses de la DB
	//timelines.DropInterests (*session)
	// Borramos las categorias de personalidad de DB
	//timelines.DropCategories (*session)
	
	/*
	
	// Convertimos la lista de intereses sacada de BD a Map
	interestsMap := timelines.ListsToMapsInterests (interesesOrdenado2)

	// Convertimos la lista de categorias sacada de BD a map
	categoriesMap := timelines.ListsToMapsCategories (personalidadesOrdenado2)
	
	*/
	// interesesOrdenado3 := timelines.GetInterestsFromDB (*session)

	// return
	// resultados := timelines.GetDatastoreResults()
	// interestsWords, personalityWords := timelines.TimelinesToMaps(*session, resultados)
	// interesesOrdenado := timelines.MapToListInterests(interestsWords)
	// personalidadesOrdenado := timelines.MapToListCategories(personalityWords)
	// timelines.InterestsToDB (interesesOrdenado, *session)
	// timelines.CategoriesToDB (personalidadesOrdenado, *session)
	// return
	// timelines.DropInterests (*session)
	// timelines.DropCategories (*session)
	// return
	// interesesOrdenado2 := timelines.GetInterestsFromDB (*session)
	// personalidadesOrdenado2 := timelines.GetCategoriesFromDB (*session)
	// return
	resultados := timelines.GetDatastoreResults()
	timelines.CalculateInterestsAndCategories(*session, resultados)
	return 
	resultados = timelines.GetDatastoreResults()
	timelines.GetAllTimelinesToDB(*session, resultados)
	return
	
	interesesOrdenado2 := timelines.GetInterestsFromDB (*session)
	personalidadesOrdenado2 := timelines.GetCategoriesFromDB (*session)
	// Convertimos la lista de intereses sacada de BD a Map
	interestsMap := timelines.ListsToMapsInterests (interesesOrdenado2)
	// Convertimos la lista de categorias sacada de BD a map
	categoriesMap := timelines.ListsToMapsCategories (personalidadesOrdenado2)
	
	//max := 0
	for i:=0; i<len(interesesOrdenado2); i++ {
		fmt.Println(timelines.Intereses[i], len(interestsMap[i]))
		// if len(interesesOrdenado2[i]) > 10 {
			// max = 10
		// } else {
			// max = len(interesesOrdenado2[i])
		// }
		// for j:=0; j<max; j++ {
			// fmt.Println(interesesOrdenado2[i][j])
		// }
	}
	fmt.Println("-----------------------")
	for i:=0; i<len(categoriesMap); i++ {
		fmt.Println(timelines.Personalidades[i], len(categoriesMap[i]))
	}	
	res := timelines.AnalyzeTimeline("GaLaeK", interestsMap, categoriesMap, *session)
	fmt.Println("-----------------------")
	// for i:=0; i<15; i++ {
		// fmt.Println(timelines.Intereses[i], res.Interests[i])
	// }
	tot1 := 0
	for i:=0; i<15; i++ {
		tot1 = tot1 + res.Interests[i]
	}
	if tot1 == 0 { fmt.Println("Tweets ocultos? o pagina borrada?"); return }
	for i:=0; i<15; i++ {
		fmt.Println(timelines.Intereses[i], "=", res.Interests[i]*100 / tot1, "%")
	}
	fmt.Println("-----------------------")
	// for i:=0; i<10; i++ {
		// fmt.Println(timelines.Personalidades[i], res.Categories[i])
	// }
	for i:=0; i<10; i = i + 2 {
		tot := res.Categories[i] + res.Categories[i+1]
		fmt.Println(timelines.Personalidades[i], "=", res.Categories[i]*100 / tot, "%")
		fmt.Println(timelines.Personalidades[i+1], "=", res.Categories[i+1]*100 / tot, "%")
	}
	
}