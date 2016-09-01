package main

import (

	"fmt"
	//"timelines"
	"labix.org/v2/mgo"	
)

func main() {
	
	// Connecting to local mongo DB
	session, err := mgo.Dial("localhost") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 

	fmt.Println("Comienza la generacion de los mapas:")
	
	// Borramos la info existente
	//timelines.DropInterests (*session)
	//timelines.DropCategories (*session)	
	//timelines.DropAges(*session)
	
	// Obtenemos los resultados de las encuestas de 'datastore'
	//resultados := timelines.GetDatastoreResults()
	
	// Obtenemos todas las timelines de las encuestas
	//timelines.GetAllTimelinesToDB(*session, resultados)
	
	// Calculamos los mapas y guardamos listas en DB
	//timelines.CalculateInterestsAndCategories(*session, resultados)
	//timelines.CalculateAges(*session, resultados)
	//timelines.CalculateGender(*session, resultados)
	
	fmt.Println("Listas guardadas en la BD correctamente")
}
