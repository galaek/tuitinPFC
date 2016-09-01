package tweel_in

import (
	"appengine"
	"appengine/urlfetch"
	"bitbucket.org/548017/go_mobydick"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

func redirecter(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/"+strings.TrimLeft(r.FormValue("user"), "/@"), 303)
}

func users(w http.ResponseWriter, r *http.Request) {
	murl := strings.TrimLeft(r.URL.Path, "/@")
	if (murl == "/") || (murl == "") {
		//	inicio(w, r)
		return
	}
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	go_mobydick.SetConsumerKey(twitterConsumerKey)
	go_mobydick.SetConsumerSecret(twitterConsumerSecret)
	api := go_mobydick.NewTwitterApi(twitterAccessToken, twitterAccessTokenSecret)

	v := url.Values{}
	v.Set("count", "1")
	v.Set("screen_name", murl)

	result, err := api.GetUserTimeline(v)
	if err != nil {
		c.Errorf("%v", err)
	}
	var tmpl *template.Template
	if len(result) == 0 {
		tmpl, err = template.ParseFiles("templates/base.html", "templates/no-tweet.html")
		if err != nil {
			c.Errorf("%v", err)
		}
		tmpl.Execute(w, nil)

	} else {

		//switch animo := analyze(result[0]); animo {
		switch animo := analyze2(result[0]); animo {
		//switch animo := valorar(result[0]); animo.Ponderacion {
		case "POSITIVA":
			tmpl, err = template.ParseFiles("templates/base.html", "templates/user-happy.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		case "NEGATIVA":
			tmpl, err = template.ParseFiles("templates/base.html", "templates/user-sad.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		default:
			tmpl, err = template.ParseFiles("templates/base.html", "templates/user.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		}

		tmpl.Execute(w, result[0])
	}
}
