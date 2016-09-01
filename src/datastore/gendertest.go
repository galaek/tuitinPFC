package main

import (

	"fmt"
	"timelines"
	"profiles"
	"strings"
	"labix.org/v2/mgo"
)


type GenderScore struct {
	Percent		int
	Hit			bool
}

func CompareGender(result [2]int, gender string) (int, bool) {
	index := 0
	var hit bool = false
	if result[0] >= result[1] {
		index = 0
	} else if result[0] < result[1] {
		index = 1
	}

	res := 0
	if gender == "Hombre" {
		res = result[0]
		if index == 0 {
			hit = true
		}
	} else if gender == "Mujer"  {
		res = result[1]
		if index == 1 {
			hit = true
		}
	}
	return res, hit
}

func CompareGenderize(genderize string, prob int, gender string) (int, bool) {
	res := 0
	hit := false
	if (genderize == "male") && (gender == "Hombre") {
		hit = true
		res = prob
	}
	if (genderize == "female") && (gender == "Mujer") {
		hit = true
		res = prob
	}

	return res, hit
}

// PrintGeneroResults prints the verification for the categories
func PrintGeneroResults (porcentajes []GenderScore) {
	totalHits := 0
	total := 0
	hitsOver70 := 0
	hitsOver60 := 0
	for i:=0; i<len(porcentajes); i++ {
		total++	
		if porcentajes[i].Hit == true {
			totalHits++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 70){
			hitsOver70++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 60){
			hitsOver60++
		}
	}
	fmt.Println("-----------------------------")
	fmt.Println("Verificacion de generos:")
	fmt.Println("-----------------------------")
	fmt.Println("Aciertos:", totalHits, "de", total, "=", (totalHits*100)/total, "%")
	fmt.Println("Aciertos +60%:", hitsOver60, "de", total, "=", (hitsOver60*100)/total, "%")
	fmt.Println("Aciertos +70%:", hitsOver70, "de", total, "=", (hitsOver70*100)/total, "%")
}

func main() {

	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	fmt.Println("Iniciando gendertest")
	resultados := timelines.GetDatastoreResults()
	//timelines.DropGender(*session)
	//timelines.CalculateGender(*session, resultados)
	genderList := timelines.GetGenderFromDB (*session)
	genderMaps := timelines.ListsToMapsGender (genderList)
	screenName := "GaLaeK"
	res := timelines.AnalyzeGenderTimeline(screenName, genderMaps, *session)
	res2 := timelines.NormalizeGenderResults(res)
	fmt.Println("Perfil:",screenName)
	timelines.PrintGenderResults(res2)
	
	// usamos genderize...
	user1 := profiles.GetUserByScreenName(screenName)
	gender, probability := DetermineGender(user1.Name)
	fmt.Println("Genderize API:", gender, "con", probability, "% de probabilidad")
	//return
	
	// Verification
	var namesMap = map[string]timelines.SurveyResult{}
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
	}
	var porcentajes []GenderScore
	var porcentajesApi []GenderScore
	var temp GenderScore
	names, _ := session.DB("analizador").CollectionNames()
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		res := timelines.AnalyzeGenderExistingTimeline(names[i], genderMaps, *session)
		res2 := timelines.NormalizeGenderResults(res)
		temp.Percent, temp.Hit = CompareGender(res2, namesMap[names[i]].Gender)
		porcentajes = append(porcentajes, temp)
		user1 = profiles.GetUserByScreenName(names[i])
		gender, probability := profiles.DetermineGender(user1.Name)
		temp.Percent, temp.Hit = CompareGenderize(gender, probability, namesMap[names[i]].Gender)
		porcentajesApi = append(porcentajesApi, temp)
	}
	PrintGeneroResults(porcentajes)
	PrintGeneroResults(porcentajesApi)
	return
}




