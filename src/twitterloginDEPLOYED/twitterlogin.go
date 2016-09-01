package manunewtesting

import (
	"appengine"
    "appengine/urlfetch"
	//"crypto/rand"
	//"encoding/base64"
	"flag"
	"fmt"
	"github.com/kurrik/oauth1a"
	//"io"
	"log"
	"net/http"
	"os"
	"time"
	//"strings"
	"profiles"
	//"trendings"
	"labix.org/v2/mgo"
	"html/template"


)

var (
		ManuKey = ""
		ManuSecret = ""
	)

var (
	settings *Settings
	service  *oauth1a.Service
	
)

// func NewSessionID() string {
	// c := 128
	// b := make([]byte, c)
	// n, err := io.ReadFull(rand.Reader, b)
	// if n != len(b) || err != nil {
		// panic("Could not generate random number")
	// }
	// return base64.URLEncoding.EncodeToString(b)
// }

func GetSessionID(req *http.Request) (id string, err error) {
	var c *http.Cookie
	if c, err = req.Cookie("session_id"); err != nil {
		return
	}
	id = c.Value
	return
}

func GetTempID(req *http.Request) (id string, err error) {
	var c *http.Cookie
	if c, err = req.Cookie("temp_id"); err != nil {
		return
	}
	id = c.Value
	return
}

func SessionstartCookie(id string) *http.Cookie {
	return &http.Cookie{
		Name:   "session_id",
		Value:  id,
		MaxAge: 900,
		Secure: false,
		Path:   "/",
	}
}

func TempStartCookie(id string) *http.Cookie {
	return &http.Cookie{
		Name:   "temp_id",
		Value:  id,
		MaxAge: 900,
		Secure: false,
		Path:   "/",
	}
}

func SessionUpdateCookie(id string) *http.Cookie {
	return &http.Cookie{
		Name:   "session_id",
		Value:  id,
		MaxAge: 900,
		Secure: false,
		Path:   "/",
	}
}

func TempEndCookie() *http.Cookie {
	return &http.Cookie{
		Name:   "temp_id",
		Value:  "",
		Secure: false,
		Path:   "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
}

func SessionEndCookie() *http.Cookie {
	return &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Secure: false,
		Path:   "/",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	}
}

func BaseHandler(rw http.ResponseWriter, req *http.Request) {
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(rw, "<a href=\"/test\">Test</a></br>")
	if sessionID, err := GetSessionID(req); err != nil {
		fmt.Fprintf(rw, "Sesión NO iniciada</br>")
		fmt.Fprintf(rw, "<a href=\"/signin\">Sign in</a></br>")
	} else {
		//fmt.Fprintf(rw, sessionID)
		if  _, prs := profiles.Sessions[sessionID]; !prs{
			fmt.Fprintf(rw, "Sesión NO iniciada</br>")
			fmt.Fprintf(rw, "<a href=\"/signin\">Sign in</a></br>")
		} else {
			http.SetCookie(rw, SessionUpdateCookie(sessionID))
			fmt.Fprintf(rw, "Sesión iniciada como " + profiles.Sessions[sessionID].AccessValues.Get("screen_name") + "</br>")
			for key, value := range profiles.Sessions[sessionID].AccessValues {
				fmt.Fprintf(rw, "Parámetros guardados " + key + " = " + value[0] + "</br>")
			}
			fmt.Fprintf(rw, "<a href=\"/timelines\">Get Timelines</a></br>")
			fmt.Fprintf(rw, "<a href=\"/logout\">Log Out</a></br>")
		}
	}
}

func TestHandler(rw http.ResponseWriter, req *http.Request) {
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	//fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
	// c := appengine.NewContext(req)
	// http.DefaultClient = urlfetch.Client(c)
    // httpClient := urlfetch.Client(c)
	// resp, err := httpClient.Get("http://www.google.com/")
    // if err != nil {
        // http.Error(rw, err.Error(), http.StatusInternalServerError)
        // return
    // }
	// var guestbookTemplate = template.Must(template.New("book").Parse(`
// <html>
  // <head>
    // <title>Test1</title>
  // </head>
  // <body>
		
  // </body>
// </html>
// `))

	// if err := guestbookTemplate.Execute(rw, nil); err != nil {
		// http.Error(rw, err.Error(), http.StatusInternalServerError)
    // }
	if tmpl, err := template.ParseFiles("templates/test1.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
    } else {
		tmpl.Execute(rw, nil)
	}
	return

}

func SignInHandler(rw http.ResponseWriter, req *http.Request) {
	var (
		url       string
		err       error
		//sessionID string
	)
	c := appengine.NewContext(req)
	http.DefaultClient = urlfetch.Client(c)
    httpClient := urlfetch.Client(c)

	//httpClient := new(http.Client)
	userConfig := &oauth1a.UserConfig{}
	if err = userConfig.GetRequestToken(service, httpClient); err != nil {
		log.Printf("Could not get request token: %v", err)
		http.Error(rw, "Problem getting the request token", 500)
		return
	}
	if url, err = userConfig.GetAuthorizeURL(service); err != nil {
		log.Printf("Could not get authorization URL: %v", err)
		http.Error(rw, "Problem getting the authorization URL", 500)
		return
	}
	// url = strings.Join([]string{url, "&force_login=true"}, "")
	// fmt.Println("-->", url, "<--")
	log.Printf("Redirecting user to %v\n", url)
	//sessionID = NewSessionID()
	log.Printf("Starting temp session %v\n", userConfig.RequestTokenKey)
	profiles.Sessions[userConfig.RequestTokenKey] = userConfig

	http.SetCookie(rw, TempStartCookie(userConfig.RequestTokenKey))
	http.Redirect(rw, req, url, 302)

}

func CallbackHandler(rw http.ResponseWriter, req *http.Request) {
	var (
		err        error
		token      string
		verifier   string
		sessionID  string
		userConfig *oauth1a.UserConfig
		ok         bool
	)
	//fmt.Println("Dem request en callback:", req.FormValue("oauth_token"))
	
	c := appengine.NewContext(req)
	http.DefaultClient = urlfetch.Client(c)
    httpClient := urlfetch.Client(c)
	log.Printf("Callback hit. %v current Sessions.\n", len(profiles.Sessions))
	if sessionID, err = GetTempID(req); err != nil {
		log.Printf("Got a callback with no session id: %v\n", err)
		http.Error(rw, "No session found", 400)
		return
	} 
	if userConfig, ok = profiles.Sessions[sessionID]; !ok {
		log.Printf("Could not find user config in sesions storage.")
		http.Error(rw, "Invalid session", 400)
		return
	}
	// The session is Correct
	if token, verifier, err = userConfig.ParseAuthorize(req, service); err != nil {
		log.Printf("Could not parse authorization: %v", err)
		http.Error(rw, "Problem parsing authorization", 500)
		return
	}

	//httpClient := new(http.Client)
	if err = userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
		log.Printf("Error getting access token: %v", err)
		http.Error(rw, "Problem getting an access token", 500)
		return
	}
	// log.Printf("Ending session %v.\n", sessionID)
	// delete(profiles.Sessions, sessionID)
	// http.SetCookie(rw, SessionEndCookie())
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	//http.SetCookie(rw, SessionstartCookie(sessionID))
	// fmt.Fprintf(rw, "<pre>")
	// fmt.Fprintf(rw, "Access Token: %v\n", userConfig.AccessTokenKey)
	// fmt.Fprintf(rw, "Token Secret: %v\n", userConfig.AccessTokenSecret)
	// fmt.Fprintf(rw, "Screen Name:  %v\n", userConfig.AccessValues.Get("screen_name"))
	// fmt.Fprintf(rw, "User ID:      %v\n", userConfig.AccessValues.Get("user_id"))
	// fmt.Fprintf(rw, "</pre>")
	// fmt.Fprintf(rw, "<a href=\"/signin\">Sign in again</a></br>")
	// fmt.Fprintf(rw, "<a href=\"/timelines\">Go get timelines kid</a>")
	
	delete(profiles.Sessions, sessionID)
	profiles.Sessions[userConfig.AccessTokenKey] = userConfig
	http.SetCookie(rw, TempEndCookie())
	http.SetCookie(rw, SessionstartCookie(userConfig.AccessTokenKey))
	http.Redirect(rw, req, "/", 302)
	
}

func LogOutHandler(rw http.ResponseWriter, req *http.Request) {
	var (
		//sessionID 	string
		err 		error
	)
	if _, err = GetSessionID(req); err != nil {
		log.Printf("Got a logout with no session id: %v\n", err)
		http.Error(rw, "Not logged in or session expired", 400)
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	fmt.Println("Sesion cerrada.\n")
	//delete(Sessions, sessionID)
	fmt.Println("Hay", len(profiles.Sessions), "sesiones registradas")
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	http.SetCookie(rw, SessionEndCookie())
	http.Redirect(rw, req, "/", 302)
}

type TimelineUser struct {
	Name string
}

func TimelinesHandler(rw http.ResponseWriter, req *http.Request) {
	var (
	//	url       string
		err       error
		sessionID string
	)
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(rw, "PrintSoemthingASAP</br>")
	//fmt.Println(req.Cookies())
	sessionID, err = GetSessionID(req)
	//if _, err = GetSessionID(req); err != nil {
	if err != nil {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		http.Error(rw, "Log In before doing stuff", 400)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	if  _, prs := profiles.Sessions[sessionID]; !prs {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	sessionID, err = GetSessionID(req)
	if sessionID == "" {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return 
	}
	t, err := template.New("encabezado").Parse("Let's read some timelines!</br>")
	err = t.ExecuteTemplate(rw, "encabezado", nil)
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
	fmt.Fprintf(rw, "Sesion ID: " + sessionID + "</br>")
	//rw.(http.Flusher).Flush()
	// dur, _ := time.ParseDuration("20s")
	// time.Sleep(dur)
	// fmt.Fprintf(rw, "After waiting...</br>")
	go GetUserTimeline(sessionID)
	fmt.Println("Proceso de captura de datos lanzado")
	return
}

// GetUserTimeline gets the user timeline to: timelinesEncuestas.screen_name 
// 						And adds the screen name to: userslist.users
func GetUserTimeline(sessionID string) {
		// Connecting to mongo DB
		session, err := mgo.Dial("130.206.83.133") 
		//session, err := mgo.Dial("localhost") 
		if err != nil { 
			panic(err) 
		} 
		defer session.Close() 
		var user string
		user = profiles.Sessions[sessionID].AccessValues.Get("screen_name")
		//user = "manupruebas"
		// names, err := session.DB(user).CollectionNames()
		// if err != nil {
			// fmt.Println("Error:", err)
		// } else {
			// fmt.Println("HERE THE NAMES:", names)
		// }
		usersColec := session.DB("userslist").C("users")

		var users []TimelineUser
		err = usersColec.Find(nil).All(&users)
		if err != nil { 
			panic(err) 
		} 
		fmt.Println("Long:", len(users), "list:", users)
		alreadyCaptured := false
		for i:=0; i<len(users); i++ {
			if users[i].Name == user {
				alreadyCaptured = true
			} 
		}
		
		if alreadyCaptured {
			fmt.Println("Timeline from", user, "already captured")
			return
		} else {
			var user1 TimelineUser 
			user1.Name = user
			err = usersColec.Insert(user1) 
			if err != nil { 
				panic(err) 
			}
			c := session.DB("timelinesEncuestas").C(user)
			profiles.GetTimelineToDBcreds(*c, user, 500, sessionID)
			/*profiles.GetTimelineToDB(*c, user, 500)*/
			// todosLosTweets := trendings.GetTweetsAll(*c)
			// for i:=0; i<len(todosLosTweets); i++ {
				// fmt.Fprintf(rw, todosLosTweets[i].Text + "</br>")
			// }
			// rw.(http.Flusher).Flush()
			// err = c.DropCollection()
			// if err != nil { 
				// panic(err) 
			// }	
			// fmt.Fprintf(rw, "</br>")
			// profiles.GetTimelineToDBcreds(*c, user, 500, sessionID)
			// /*profiles.GetTimelineToDB(*c, user, 500)*/
			// todosLosTweets2 := trendings.GetTweetsAll(*c)
			// for i:=0; i<len(todosLosTweets2); i++ {
				// fmt.Fprintf(rw, todosLosTweets2[i].Text + "</br>")
			// }
			fmt.Println("Timeline from", user, "successfully captured")
		}
}

type Settings struct {
	Key  string
	Sec  string
	Port int
}

func init() {
	profiles.Sessions = map[string]*oauth1a.UserConfig{}
	settings = &Settings{}
	flag.IntVar(&settings.Port, "port", 8080, "Port to run on")
	flag.StringVar(&settings.Key, "key", "Cij69U0URS7d0dMNkw9p2nfJp", "Consumer key of your app")
	flag.StringVar(&settings.Sec, "secret", "gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z", "Consumer secret of your app")
	flag.Parse()
	if settings.Key == "" || settings.Sec == "" {
		fmt.Fprintf(os.Stderr, "You must specify a consumer key and secret.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	service = &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authorize",
		//AuthorizeURL: "https://api.twitter.com/oauth/authenticate",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    settings.Key,
			ConsumerSecret: settings.Sec,
			//CallbackURL:    "http://localhost:8080/callback/",
			CallbackURL:    "http://manunewtesting.appspot.com/callback/",
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}
	//http.Handle("/followers/", &authHandler{handler: serveFollowers})
	http.HandleFunc("/", BaseHandler)
	http.HandleFunc("/test/", TestHandler)
	http.HandleFunc("/signin/", SignInHandler)
	http.Handle("/logout/", &authHandler{handler: LogOutHandler})
	http.HandleFunc("/callback/", CallbackHandler)
	http.Handle("/timelines/", &authHandler{handler: TimelinesHandler})
	//log.Printf("Visit http://localhost:%v in your browser\n", settings.Port)
	//log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", settings.Port), nil)) 
}

// authHandler reads the auth cookie and invokes a handler with the result.
type authHandler struct {
	handler  func(w http.ResponseWriter, r *http.Request)
	optional bool
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c, _ := r.Cookie("session_id"); c != nil {
		fmt.Println("Sesión encontrada:", c.Value)
		http.SetCookie(w, SessionUpdateCookie(c.Value))
		
	} else {
		//w.Header().Set("Content-Type", "text/html;charset=utf-8")
		//http.Error(w, "Not logged in or session expired", 400)
		fmt.Fprintf(w, "Not logged in or session expired</br>")
		fmt.Fprintf(w, "<a href=\"/\">Back to main</a></br>")
		return
	}
	h.handler(w, r)
}