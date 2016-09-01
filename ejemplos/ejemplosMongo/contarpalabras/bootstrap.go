package main

import (
   // "fmt"
    "net/http"
	"html/template"
)

type persona struct {
	Name		string
	Lastname	string
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprint(w, "Hola View!!!")
	var p1 persona
	p1.Name = "Manuelo"
	p1.Lastname = "Eyesclosed"
	t, _ := template.ParseFiles("web/starter-template/index.html")
    t.Execute(w, p1)
}

func main() {
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/golang/ejemplomongo/contarpalabras/web/"))))
	http.HandleFunc("/view/", myHandler)
    http.ListenAndServe(":8080", nil)
}