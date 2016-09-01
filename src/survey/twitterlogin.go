package manunewtesting
import _ "appengine/remote_api"
import (
	"appengine"
    "appengine/urlfetch"
	"appengine/datastore"
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
	//"labix.org/v2/mgo"
	"html/template"
	"strconv"
	"encoding/json"


)

var (
	settings *Settings
	service  *oauth1a.Service
	
)

type UserData struct {
	ScreenName	string
	UserID		int64
	AccessKey	string
	SecretKey	string
}
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

func SessionStartCookie(id string) *http.Cookie {
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
	var data UserData
	sessionID, err := GetSessionID(req)
	if err != nil {
		// Not logged in
		if tmpl, err := template.ParseFiles("html/cover.html"); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		} else {
			tmpl.Execute(rw, nil)
		}		
	} else {
		// Session found
		key, err := datastore.DecodeKey(sessionID)
		if err != nil {
			fmt.Fprintf(rw, err.Error())
		} else {
			c := appengine.NewContext(req)
			err = datastore.Get(c,key, &data)
			if err != nil {
				fmt.Fprintf(rw, err.Error())
			} else {
				// Session correct
				http.SetCookie(rw, SessionUpdateCookie(sessionID))
				if tmpl, err := template.ParseFiles("html/cover.html"); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				} else {
					tmpl.Execute(rw, nil)
				}	
			}
		}
	}
}

func SignInHandler(rw http.ResponseWriter, req *http.Request) {
	var (
		url       string
		err       error
		//sessionID string
	)
	service = &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authorize",
		//AuthorizeURL: "https://api.twitter.com/oauth/authenticate",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    settings.Key,
			ConsumerSecret: settings.Sec,
			//CallbackURL:    "http://localhost:8080/callback/",
			CallbackURL:		"http://" + req.Host + "/callback/",
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}
	
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
	//http.SetCookie(rw, SessionStartCookie(sessionID))
	// fmt.Fprintf(rw, "<pre>")
	// fmt.Fprintf(rw, "Access Token: %v\n", userConfig.AccessTokenKey)
	// fmt.Fprintf(rw, "Token Secret: %v\n", userConfig.AccessTokenSecret)
	// fmt.Fprintf(rw, "Screen Name:  %v\n", userConfig.AccessValues.Get("screen_name"))
	// fmt.Fprintf(rw, "User ID:      %v\n", userConfig.AccessValues.Get("user_id"))
	// fmt.Fprintf(rw, "</pre>")
	// fmt.Fprintf(rw, "<a href=\"/signin\">Sign in again</a></br>")
	// fmt.Fprintf(rw, "<a href=\"/timelines\">Go get timelines kid</a>")
	
	delete(profiles.Sessions, sessionID)
	userID, _ := strconv.ParseInt(userConfig.AccessValues.Get("user_id"), 10, 64)
	data := UserData{
			ScreenName: 	userConfig.AccessValues.Get("screen_name"),
			UserID:			userID,
			AccessKey: 		userConfig.AccessTokenKey,
			SecretKey:		userConfig.AccessTokenSecret,
	}
	
	//profiles.Sessions[userConfig.AccessTokenKey] = userConfig
	// We store the session in the datastore
	key := datastore.NewKey(c, "data", userConfig.AccessTokenKey, 0, nil)
	//key := datastore.NewIncompleteKey(c, "data", nil)
	key, err = datastore.Put(c, key, &data)
	if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
	} 
	stringedKey := key.Encode()
	http.SetCookie(rw, TempEndCookie())
	//http.SetCookie(rw, SessionStartCookie(userConfig.AccessTokenKey))
	http.SetCookie(rw, SessionStartCookie(stringedKey))
	http.Redirect(rw, req, "/survey/", 302)
	
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

func TwitterCardHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/twittercard.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}

}

func ResultadosManuHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	c := appengine.NewContext(req)
	q := datastore.NewQuery("surveyresult").KeysOnly()
	// var res []SurveyResult
	// q.GetAll(c, &res)
	// fmt.Fprintf(rw, "Num results: %d<br>",(len(res)))
	// for i:=0; i<len(res); i++ {
	
		// fmt.Fprintf(rw, "%s<br>",ResultToString(res[i]))
	// }
	keys, _ := q.GetAll(c, nil)
	fmt.Fprintf(rw, "Num keys: %d<br>",(len(keys)))
	for i:=0; i<len(keys); i++ {
	
		//fmt.Fprintf(rw, "%d<br>",keys[i].IntID())
	}

}

type TimelineUser struct {
	Name string
}

func SurveyHandler(rw http.ResponseWriter, req *http.Request) {
	var (
	//	url       string
		err       error
		sessionID string
	)
	var data UserData
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	//fmt.Println(req.Cookies())
	sessionID, err = GetSessionID(req)
	if err != nil {
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder hacer la super mega encuesta</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	
	key, err := datastore.DecodeKey(sessionID)
	if err != nil {
		fmt.Fprintf(rw, err.Error() + "</br>")
		fmt.Fprintf(rw, "DecodeKey failed...</br>")
	} else {
		c := appengine.NewContext(req)
		err = datastore.Get(c,key, &data)
		if err != nil {
			fmt.Fprintf(rw, err.Error() + "</br>")
			fmt.Fprintf(rw, "Could not get data from datastore...</br>")
			fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
			return
		} else {
			http.SetCookie(rw, SessionUpdateCookie(sessionID))
				if tmpl, err := template.ParseFiles("html/survey.html"); err != nil {
					http.Error(rw, err.Error(), http.StatusInternalServerError)
				} else {
					tmpl.Execute(rw, &data)
				}
			//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
			//go GetUserTimeline(sessionID)
			//fmt.Fprintf(rw, "Proceso de captura de datos lanzado")
		}
	}
	return
}

func ProfilesCalculator (result SurveyResult) []int {
	scores := make([]int, 8)
	// Twittero Zen
	scores[0] = ((result.Personal[1] + result.Situation[1])*5 + (result.Personal[2] + 6-result.Situation[2])*3 +
				(result.Personal[4] + 6-result.Situation[4])*2 + (6-result.Personal[3] + result.Situation[3]))
	// Twittero Loco
	scores[1] = ((result.Personal[0] + 6-result.Situation[0])*5 + (6-result.Personal[1] + 6-result.Situation[1])*3 +
				(6-result.Personal[4] + result.Situation[4])*2 + (result.Personal[3] + 6-result.Situation[3]))		
	// Twittero Espia
	scores[2] = ((6-result.Personal[0] + result.Situation[0])*5 + (result.Personal[1] + result.Situation[1])*3 +
				(6-result.Personal[3] + result.Situation[3])*2 + (result.Personal[4] + 6-result.Situation[4]))		
	// Twittero Negativo
	scores[3] = ((6-result.Personal[2] + result.Situation[2])*5 + (6-result.Personal[0] + result.Situation[0])*3 +
				(result.Personal[1] + result.Situation[1])*2 + (6-result.Personal[3] + result.Situation[3]))	
	// Twittero Lider
	scores[4] = ((result.Personal[3] + 6-result.Situation[3])*5 + (result.Personal[0] + 6-result.Situation[0])*3 +
				(result.Personal[2] + 6-result.Situation[2])*2 + (result.Personal[1] + result.Situation[1]))
	// Twittero Gruñon
	scores[5] = ((6-result.Personal[2] + result.Situation[2])*5 + (6-result.Personal[1] + 6-result.Situation[1])*3 +
				(result.Personal[0] + 6-result.Situation[0])*2 + (result.Personal[3] + 6-result.Situation[3]))	
	// Twittero Ordenado
	scores[6] = ((result.Personal[4] + 6-result.Situation[4])*5 + (result.Personal[1] + result.Situation[1])*3 +
				(6-result.Personal[0] + result.Situation[0])*2 + (6-result.Personal[3] + result.Situation[3]))			
	// Twittero Optimista
	scores[7] = ((result.Personal[2] + 6-result.Situation[2])*5 + (result.Personal[0] + 6-result.Situation[0])*3 +
				(result.Personal[1] + result.Situation[1])*2 + (result.Personal[3] + 6-result.Situation[3]))						
	return scores
}	

func SelectProfile (scores []int) int {
	result := 0
	maxScore := 0
	for i:=0; i<8; i++ {
		if scores[i] > maxScore {
			maxScore = scores[i]
			result = i
		}
	}
	return result
}

type ProfileInfo struct {
	Nombre		string
	Descripción	string
	Hashtag		string
	ImagenURL	string
	ScreenName	string
	NombrePlano	string
}

func ResultToString (res SurveyResult) string {
	out, _ := json.Marshal(res)
	return string(out)
}

type SurveyResult struct {
	ScreenName	string
	Born		string
	Country		string
	Gender		string
	Personal	[]int
	Situation	[]int
	Interests	[]string
	Scores		[]int
	Profile		int
}
func ResultsHandler(rw http.ResponseWriter, req *http.Request) {
	var (
	//	url       string
		err       error
		sessionID string
	)
	//result := new(SurveyResult)
	result := SurveyResult {
		ScreenName:	"",
		Born: 		"",
		Country:	"",
		Gender:		"",
		Personal:	make([]int, 5),
		Situation:	make([]int, 5),
		Scores:		make([]int, 8),
	}
	
	//var result SurveyResult
	result.Country = req.FormValue("country")
	result.Gender = req.FormValue("gender")
	result.Born = req.FormValue("borndate")
	result.Personal[0], _ = strconv.Atoi(req.FormValue("pers1"))
	result.Personal[1], _ = strconv.Atoi(req.FormValue("pers2"))
	result.Personal[2], _ = strconv.Atoi(req.FormValue("pers3"))
	result.Personal[3], _ = strconv.Atoi(req.FormValue("pers4"))
	result.Personal[4], _ = strconv.Atoi(req.FormValue("pers5"))
	result.Situation[0], _ = strconv.Atoi(req.FormValue("situacion1"))
	result.Situation[1], _ = strconv.Atoi(req.FormValue("situacion2"))
	result.Situation[2], _ = strconv.Atoi(req.FormValue("situacion3"))
	result.Situation[3], _ = strconv.Atoi(req.FormValue("situacion4"))
	result.Situation[4], _ = strconv.Atoi(req.FormValue("situacion5"))
	
	if err := req.ParseForm(); err != nil {
	// handle error
	}
	result.Interests = req.Form["intereses[]"]
	
	result.Scores = ProfilesCalculator(result)
	result.Profile = SelectProfile(result.Scores)
	
	var data UserData
	//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	//fmt.Println(req.Cookies())

	sessionID, err = GetSessionID(req)
	if err != nil {
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder hacer la super mega encuesta</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	
	key, err := datastore.DecodeKey(sessionID)
	if err != nil {
		fmt.Fprintf(rw, err.Error() + "</br>")
		fmt.Fprintf(rw, "DecodeKey failed...</br>")
	} else {
		c := appengine.NewContext(req)
		profilesQuery := datastore.NewQuery("perfil")
		var profiles []ProfileInfo
		profilesQuery.GetAll(c, &profiles)
		
		err = datastore.Get(c,key, &data)
		if err != nil {
			fmt.Fprintf(rw, err.Error() + "</br>")
			fmt.Fprintf(rw, "Could not get data from datastore...</br>")
			fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
			return
		} else {
			//http.SetCookie(rw, SessionUpdateCookie(sessionID))
			result.ScreenName = data.ScreenName
			profiles[result.Profile].ScreenName = result.ScreenName
			
			profilesQuery := datastore.NewQuery("surveyresult")
			var allresults []SurveyResult
			profilesQuery.GetAll(c, &allresults)
			alreadyIn := false
			for i:=0; i<len(allresults); i++ {
				if allresults[i].ScreenName == result.ScreenName {
					alreadyIn = true
					break
				}
			}
			if alreadyIn == false {	
				// New entry, We store the "result" struct in DataStore
				_, err := datastore.Put(c, datastore.NewIncompleteKey(c, "surveyresult", nil), &result)
				if err != nil {
					result.ScreenName = err.Error()
				}	
				//rw.Header().Set("Content-Type", "text/html;charset=utf-8")
				//fmt.Fprintf(rw, "No sta en la lista %s<br>", result.ScreenName)
			} 
			
			//http.SetCookie(rw, SessionEndCookie())
			if tmpl, err := template.ParseFiles("html/result.html"); err != nil {
				http.Error(rw, err.Error(), http.StatusInternalServerError)
			} else {
				tmpl.Execute(rw, &profiles[result.Profile])
			}
		}
	}
	return
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
	// flag.StringVar(&settings.Key, "key", "Cij69U0URS7d0dMNkw9p2nfJp", "Consumer key of your app")
	// flag.StringVar(&settings.Sec, "secret", "gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z", "Consumer secret of your app")
	flag.StringVar(&settings.Key, "key", "wU0h2rYWxM3X6S7oJUP4Osy4N", "Consumer key of your app")
	flag.StringVar(&settings.Sec, "secret", "8iuxhG4IOVxbQUmiYJnzkyn6dYznwTzxO5QLhrO8u91ZxAteFr", "Consumer secret of your app")
	flag.Parse()
	if settings.Key == "" || settings.Sec == "" {
		fmt.Fprintf(os.Stderr, "You must specify a consumer key and secret.\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	


	
	//http.Handle("/followers/", &authHandler{handler: serveFollowers})
	http.HandleFunc("/", BaseHandler)
	http.HandleFunc("/signin/", SignInHandler)
	//http.HandleFunc("/twittercard/", TwitterCardHandler)
	http.HandleFunc("/loco/", LocoHandler)
	http.HandleFunc("/zen/", ZenHandler)
	http.HandleFunc("/espia/", EspiaHandler)
	http.HandleFunc("/negativo/", NegativoHandler)
	http.HandleFunc("/lider/", LiderHandler)
	http.HandleFunc("/ordenado/", OrdenadoHandler)
	http.HandleFunc("/grunon/", GrunonHandler)
	http.HandleFunc("/optimista/", OptimistaHandler)
	http.Handle("/logout/", &authHandler{handler: LogOutHandler})
	http.HandleFunc("/callback/", CallbackHandler)
	http.Handle("/survey/", &authHandler{handler: SurveyHandler})
	http.Handle("/results/", &authHandler{handler: ResultsHandler})
	http.HandleFunc("/resultadosmanu/", ResultadosManuHandler)
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
		http.Redirect(w, r, "/", 302)
		// fmt.Fprintf(w, "Not logged in or session expired</br>")
		// fmt.Fprintf(w, "<a href=\"/\">Back to main</a></br>")
		return
	}
	h.handler(w, r)
}

func LocoHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/loco.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func ZenHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/zen.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func EspiaHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/espia.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func NegativoHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/negativo.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func LiderHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/lider.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func GrunonHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/grunon.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func OrdenadoHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/ordenado.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}
func OptimistaHandler(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-Type", "text/html;charset=utf-8")
	if tmpl, err := template.ParseFiles("html/optimista.html"); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	} else {
		tmpl.Execute(rw, nil)
	}
}