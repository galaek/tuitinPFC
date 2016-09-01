package main

import (

	"fmt"
	"timelines"
	"profiles"
	"labix.org/v2/mgo"	
)

func main() {
	
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	fmt.Println("Comienza el analisis:")
	screenName := "JavierSantiso"
	// Leemos los mapas de la BD
	interesesList := timelines.GetInterestsFromDB (*session)
	personalidadesList := timelines.GetCategoriesFromDB (*session)
	agesList := timelines.GetAgesFromDB (*session)
	genderList := timelines.GetGenderFromDB (*session)
	
	interestsMap := timelines.ListsToMapsInterests (interesesList)
	categoriesMap := timelines.ListsToMapsCategories (personalidadesList)
	agesMap := timelines.ListsToMapsAges (agesList)	
	genderMaps := timelines.ListsToMapsGender (genderList)
	
	res := timelines.AnalyzeTimeline(screenName, interestsMap, categoriesMap, agesMap, genderMaps, *session)
	res2 := timelines.NormalizeResults(res)
	
	fmt.Println("----------------------")
	fmt.Println("Perfil:",screenName)
	fmt.Println("----------------------")
	timelines.PrintResults(res2)
	user1 := profiles.GetUserByScreenName(screenName)
	gender, probability := profiles.DetermineGender(user1.Name)
	fmt.Println("Genderize API:", gender, "con", probability, "% de probabilidad")

	
}
