package manunewtesting

import _ "appengine/remote_api"

import (
        "html/template"
        "net/http"
        "time"
		"fmt"
        "appengine"
        "appengine/datastore"
        "appengine/user"
)

type Greeting struct {
		Author  string
        Content string
        Date    time.Time
}

func init() {
        http.HandleFunc("/", root)
        http.HandleFunc("/sign", sign)
		http.HandleFunc("/mostrarcosas/", MostrarCosas)
}

// guestbookKey returns the key used for all guestbook entries.
func guestbookKey(c appengine.Context) *datastore.Key {
        // The string "default_guestbook" here could be varied to have multiple guestbooks.
        return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        // Ancestor queries, as shown here, are strongly consistent with the High
        // Replication Datastore. Queries that span entity groups are eventually
        // consistent. If we omitted the .Ancestor from this query there would be
        // a slight chance that Greeting that had just been written would not
        // show up in a query.
        q := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
        greetings := make([]Greeting, 0, 10)
        if _, err := q.GetAll(c, &greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        if err := guestbookTemplate.Execute(w, greetings); err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

var guestbookTemplate = template.Must(template.New("book").Parse(`
<html>
  <head>
    <title>Go Guestbook</title>
  </head>
  <body>
    {{range .}}
      {{with .Author}}
        <p><b>{{.}}</b> wrote:</p>
      {{else}}
        <p>Someone or something wrote:</p>
      {{end}}
      <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
    </form>
  </body>
</html>
`))

func sign(w http.ResponseWriter, r *http.Request) {
        c := appengine.NewContext(r)
        g := Greeting{
                Content: r.FormValue("content"),
                Date:    time.Now(),
        }
        if u := user.Current(c); u != nil {
                g.Author = u.String()
        }
        // We set the same parent key on every Greeting entity to ensure each Greeting
        // is in the same entity group. Queries across the single entity group
        // will be consistent. However, the write rate to a single entity group
        // should be limited to ~1/second.
        key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
        _, err := datastore.Put(c, key, &g)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/", http.StatusFound)
}



func MostrarCosas(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	c := appengine.NewContext(req)
	q := datastore.NewQuery("Greeting").KeysOnly()
	// var res []SurveyResult
	// q.GetAll(c, &res)
	// fmt.Fprintf(rw, "Num results: %d<br>",(len(res)))
	// for i:=0; i<len(res); i++ {
	
		// fmt.Fprintf(rw, "%s<br>",ResultToString(res[i]))
	// }
	var data Greeting
	keys, _ := q.GetAll(c, nil)
	fmt.Fprintf(rw, "Num keys: %d<br>",(len(keys)))
	for i:=0; i<len(keys); i++ {
		//fmt.Println(string(keys[i]) + "<br>")
		
		datastore.Get(c,keys[i], &data)
		fmt.Fprintf(rw, "%d ",keys[i].IntID())
		fmt.Fprintf(rw, "%s ", data.Author)
		fmt.Fprintf(rw, "%s ", data.Content)
		fmt.Fprintf(rw, "%s ", data.Date)
		fmt.Fprintf(rw, "%s <br>", data)
	}

}
