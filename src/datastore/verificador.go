package main

import (

	"fmt"
	"timelines"
	//"profiles"
	"labix.org/v2/mgo"
	"time"	
)

func main() {
	
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	resultados := timelines.GetDatastoreResults()
	//return
	var namesMap = map[string]timelines.SurveyResult{}
	puntuaciones := make([]timelines.Scores, len(resultados))
	for i:=0; i<len(resultados); i++ {
		namesMap[resultados[i].ScreenName] = resultados[i]
		puntuaciones[i] = timelines.CalculateScores(resultados[i])
	}
	fmt.Println("Comienza la verificacion:")
	// Cargamos los mapas
	interesesList := timelines.GetInterestsFromDB (*session)
	personalidadesList := timelines.GetCategoriesFromDB (*session)
	agesList := timelines.GetAgesFromDB (*session)
	genderList := timelines.GetGenderFromDB (*session)
	
	interestsMap := timelines.ListsToMapsInterests (interesesList)
	categoriesMap := timelines.ListsToMapsCategories (personalidadesList)
	agesMap := timelines.ListsToMapsAges (agesList)
	genderMaps := timelines.ListsToMapsGender (genderList)
	
	// Obtenemos los nombres de los perfiles usados para calcular las listas
	names, _ := session.DB("analizador").CollectionNames()
	var porcentajesInt []int
	var aciertosCat []int
	var porcentajesAge []AgeScore
	var temp AgeScore
	var porcentajesGender []GenderScore
	// var porcentajesApi []GenderScore
	var temp2 GenderScore
	// Verification time
	for i:=0; i<len(names); i++ {
		if names[i] == "system.indexes" {continue}
		// Leemos la timeline, analizamos y normalizamos
		res := timelines.AnalyzeExistingTimeline(names[i], interestsMap, categoriesMap, agesMap, genderMaps, *session)
		res2 := timelines.NormalizeResults(res)
		
		// Calculamos el acierto de los intereses
		porcentajesInt = append(porcentajesInt, CompareInterests(Top4Interests(res2), namesMap[names[i]].Interests))		
		// Calculamos el acierto de las categorias
		aciertosCat = append(aciertosCat, CompareCategories(res2, puntuaciones[i]))
		// Calculamos el acierto de las edades
		temp.Percent, temp.Hit = CompareAges(res2.Ages, namesMap[names[i]].Born)
		porcentajesAge = append(porcentajesAge, temp)
		// Calculamos el acierto de los generos
		temp2.Percent, temp2.Hit = CompareGender(res2.Gender, namesMap[names[i]].Gender)
		porcentajesGender = append(porcentajesGender, temp2)
		// Calculamos el acierto de genderize
		// user1 := profiles.GetUserByScreenName(names[i])
		// gender, probability := profiles.DetermineGender(user1.Name)
		// temp2.Percent, temp2.Hit = CompareGenderize(gender, probability, namesMap[names[i]].Gender)
		// porcentajesApi = append(porcentajesApi, temp2)
	}
	// Show Time
	PrintInterestsResults(porcentajesInt)
	PrintCategoriesResults(aciertosCat)
	PrintEdadesResults(porcentajesAge)
	PrintGeneroResults(porcentajesGender)
	// PrintGenderizeResults(porcentajesApi)
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

func CompareCategories(res timelines.ResultScore, scores timelines.Scores) int{
	var hits int = 0
	var misses int = 0
	if (scores.Introvertido >=6 && res.Categories[0] >= 50) || (scores.Introvertido <=6 && res.Categories[0] <= 50) {
		hits++
	}
	if (scores.Introvertido >6 && res.Categories[0] < 50) || (scores.Introvertido <6 && res.Categories[0] > 50) {
		misses++
	}
	//fmt.Println(res.Categories[0], "vs", scores.Introvertido
	if (scores.Agresivo >=6 && res.Categories[2] >= 50) || (scores.Agresivo <=6 && res.Categories[2] <= 50) {
		hits++
	}
	if (scores.Agresivo >6 && res.Categories[2] < 50) || (scores.Agresivo <6 && res.Categories[2] > 50) {
		misses++
	}
	if (scores.Positivo >=6 && res.Categories[4] >= 50) || (scores.Positivo <=6 && res.Categories[4] <= 50) {
		hits++
	}
	if (scores.Positivo >6 && res.Categories[4] < 50) || (scores.Positivo <6 && res.Categories[4] > 50) {
		misses++
	}
	if (scores.Lider >=6 && res.Categories[6] >= 50) || (scores.Lider <=6 && res.Categories[6] <= 50) {
		hits++
	}
	if (scores.Lider >6 && res.Categories[6] < 50) || (scores.Lider <6 && res.Categories[6] > 50) {
		misses++
	}
	if (scores.Metodico >=6 && res.Categories[8] >= 50) || (scores.Metodico <=6 && res.Categories[8] <= 50) {
		hits++
	}
	if (scores.Metodico >6 && res.Categories[8] < 50) || (scores.Metodico <6 && res.Categories[8] > 50) {
		misses++
	}
	//fmt.Println("Aciertos:", hits, "Fallos:", misses)
	return hits
}

// PrintCategoriesResults prints the verification for the categories
func PrintCategoriesResults (aciertos []int) {
	res20 := 0
	res40 := 0
	res60 := 0
	res80 := 0
	res100 := 0
	for i:=0; i<len(aciertos); i++ {
		if aciertos[i] == 1 {
			res20++
		}
		if aciertos[i] == 2 {
			res40++
		}
		if aciertos[i] == 3 {
			res60++
		}
		if aciertos[i] == 4 {
			res80++
		}
		if aciertos[i] == 5 {
			res100++
		}
	}
	fmt.Println("-----------------------------")
	fmt.Println("Verificacion de categorias:")
	fmt.Println("-----------------------------")
	//fmt.Println("Aciertos de +40% =",  res40 + res60 + res80 + res100, "de", len(aciertos), "=",  (res40 + res60 + res80 + res100)*100/len(aciertos), "%")
	fmt.Println("Aciertos de +60% =",  res60 + res80 + res100, "de", len(aciertos), "=",  (res60 + res80 + res100)*100/len(aciertos), "%")
	fmt.Println("Aciertos de +80% =",  res80 + res100, "de", len(aciertos), "=",  (res80 + res100)*100/len(aciertos), "%")
	fmt.Println("Aciertos de 100% =", res100, "de", len(aciertos), "=", res100*100/len(aciertos), "%")
}

// PrintInterestsResults prints the verification for the interests results
func PrintInterestsResults (porcentajes []int) {
	masde50 := 0
	masde100 := 0
	masde75 := 0
	for i:=0; i<len(porcentajes); i++ {
		if porcentajes[i] >= 50 {
			masde50++
		}
		if porcentajes[i] == 100 {
			masde100++
		}
		if porcentajes[i] >= 75 {
			masde75++
		}
	}
	fmt.Println("-----------------------------")
	fmt.Println("Verificacion de intereses:")
	fmt.Println("-----------------------------")
	fmt.Println("Aciertos por encima de 50% =", masde50, "de", len(porcentajes), "=", masde50*100/len(porcentajes), "%")
	fmt.Println("Aciertos por encima de 75% =", masde75, "de", len(porcentajes), "=", masde75*100/len(porcentajes), "%")
	fmt.Println("Aciertos de 100% =", masde100, "de", len(porcentajes), "=", masde100*100/len(porcentajes), "%")
}

func CompareInterests (top4 []string, list []string) int{
	contador := 0
	for i:=0; i<4; i++ {
		for j:=0; j<len(list); j++ {
			if top4[i] == list[j] {
				contador++
			}
		}
	}
	return contador*100 / len(list)
}

func Top4Interests (res timelines.ResultScore) []string {
	top4 := make([]string, 4)
	temp := res.Interests
	for j:=0; j<4; j++ {
		maxIndex := 0
		maxValue := 0
		for i:=0; i<15; i++ {
			if temp[i] > maxValue {
				maxIndex = i
				maxValue = temp[i]
			}
		}
		temp[maxIndex] = 0
		top4[j] = timelines.Intereses[maxIndex]
	}
	return top4
}

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

// PrintEdadessResults prints the verification for the categories
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
	} else if gender == "Mujer" {
		res = result[1]
		if index == 1 {
			hit = true
		}
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

// PrintGeneroResults prints the verification for the categories
func PrintGenderizeResults (porcentajes []GenderScore) {
	totalHits := 0
	total := 0
	hitsOver90 := 0
	hitsOver80 := 0
	for i:=0; i<len(porcentajes); i++ {
		total++	
		if porcentajes[i].Hit == true {
			totalHits++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 90){
			hitsOver90++
		}
		if (porcentajes[i].Hit == true) && (porcentajes[i].Percent >= 80){
			hitsOver80++
		}
	}
	fmt.Println("-----------------------------")
	fmt.Println("Verificacion de genderize:")
	fmt.Println("-----------------------------")
	fmt.Println("Aciertos:", totalHits, "de", total, "=", (totalHits*100)/total, "%")
	fmt.Println("Aciertos +80%:", hitsOver80, "de", total, "=", (hitsOver80*100)/total, "%")
	fmt.Println("Aciertos +90%:", hitsOver90, "de", total, "=", (hitsOver90*100)/total, "%")
}