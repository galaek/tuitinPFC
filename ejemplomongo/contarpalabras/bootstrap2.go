package main

import (
   // "fmt"
    "net/http"
	"html/template"
	"os"
)

type justFilesFilesystem struct {
    fs http.FileSystem
}

func (fs justFilesFilesystem) Open(name string) (http.File, error) {
    f, err := fs.fs.Open(name)
    if err != nil {
        return nil, err
    }
    return neuteredReaddirFile{f}, nil
}

type neuteredReaddirFile struct {
    http.File
}

func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
    return nil, nil
}

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
	fs := justFilesFilesystem{http.Dir("C:/golang/ejemplomongo/contarpalabras/web/")}
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(fs)))
	//http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("C:/golang/ejemplomongo/contarpalabras/web/"))))
	http.HandleFunc("/view/", myHandler)
    http.ListenAndServe(":8080", nil)
}