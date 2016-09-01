package main

import (

	"fmt"
	"timelines"
	"labix.org/v2/mgo"	
	"time"
)


func CompareAges(result [3]int, born string) (int, bool) {
	max := 0
	index := 0
	var hit bool = false
	for i:=0; i<3; i++ {
		if result[i] > max {
			max = result[i]
			index = i
		}
	}
	res := 0
	date := timelines.StringToDate(born)
	const shortForm = "2006-Jan-02"
	t1, _ := time.Parse(shortForm, "1980-Jan-01")
	t2, _ := time.Parse(shortForm, "1990-Jan-01")
	if date.Before(t1) {
		res = result[0]
		if index == 0 {
			hit = true
		}
	} else if date.Before(t2) {
		res = result[1]
		if index == 1 {
			hit = true
		}
	} else {
		res = result[2]
		if index == 2 {
			hit = true
		}
	}
	return res, hit
}

type AgeScore struct {
	Percent		int
	Hit			bool
}

// PrintCategoriesResults prints the verification for the categories
func PrintEdadesResults (porcentajes []AgeScore) {
	totalHits := 0
	total := 0
	hitsOver40 := 0
	hitsOver50 := 0
	hitsOver60 := 0
	for i:=0; i<len(porcentajes); i++ {
		total++	
		if porcentajes[i].Hit == true {
			totalHits++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 40){
			hitsOver40++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 50){
			hitsOver50++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 60){
			hitsOver60++
		}
	}
	fmt.Println("-----------------------------")
	fmt.Println("Verificacion de edades:")
	fmt.Println("-----------------------------")
	fmt.Println("Aciertos:", totalHits, "de", total, "=", (totalHits*100)/total, "%")
	fmt.Println("Aciertos +40%:", hitsOver40, "de", total, "=", (hitsOver40*100)/total, "%")
	fmt.Println("Aciertos +50%:", hitsOver50, "de", total, "=", (hitsOver50*100)/total, "%")
	fmt.Println("Aciertos +60%:", hitsOver60, "de", total, "=", (hitsOver60*100)/total, "%")
}

func main() {

	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	resultados := timelines.GetDatastoreResults()
	//timelines.DropAges(*session)
	//timelines.CalculateAges(*session, resultados)
	agesList := timelines.GetAgesFromDB (*session)
	agesMaps := timelines.ListsToMapsAges (agesList)
	screenName := "Isamar_cachis"
	res := timelines.AnalyzeAgeTimeline(screenName, agesMaps, *session)
	res2 := timelines.NormalizeAgeResults(res)
	fmt.Println("Perfil:",screenName)
	timelines.PrintAgesResults(res2)
	
	// Verification
	var namesMap = map[string]timelines.SurveyResult{}
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
	}
	var porcentajes []AgeScore
	var temp AgeScore
	names, _ := session.DB("analizador").CollectionNames()
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		res := timelines.AnalyzeAgesExistingTimeline(names[i], agesMaps, *session)
		res2 := timelines.NormalizeAgeResults(res)
		temp.Percent, temp.Hit = CompareAges(res2, namesMap[names[i]].Born)
		porcentajes = append(porcentajes, temp)
	}
	PrintEdadesResults(porcentajes)
	return
}



