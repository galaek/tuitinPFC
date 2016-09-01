package main 
import ( 
		"fmt" 
		"labix.org/v2/mgo" 
		"labix.org/v2/mgo/bson" 
		) 

type Person struct { 
	Name string 
	Phone string 
} 

func main() { 
	session, err := mgo.Dial("localhost") 
	
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	// Optional. Switch the session to a monotonic behavior. session.SetMode(mgo.Monotonic, true) 
	c := session.DB("test").C("people") 
	err = c.Insert(&Person{"Ale", "+55 53 8116 9639"}, &Person{"Cla", "+55 53 8402 8510"}) 
	if err != nil { 
		panic(err) 
	} 
	//result := Person{} 
	var persona []Person
	err = c.Find(bson.M{"name": bson.M{"$exists": true}}).All(&persona)
	//err = c.Find(bson.M{"name": "Ale"}).All(&persona) 
	//err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil { 
		panic(err) 
	} 
	fmt.Println("Longitud: ", len(persona))
	fmt.Println("Persona: ", persona)
	fmt.Println("Phone:", persona[1].Phone) 
}