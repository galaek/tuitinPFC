package tweel_in

import (
	"appengine"
	"appengine/urlfetch"
	"bitbucket.org/548017/go_mobydick"
	"encoding/json"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/kurrik/oauth1a"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	service *oauth1a.Service
)

// authHandler reads the auth cookie and invokes a handler with the result.
type authHandler struct {
	handler  func(w http.ResponseWriter, r *http.Request, c *oauth.Credentials)
	optional bool
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var cred *oauth.Credentials
	if c, _ := r.Cookie("auth"); c != nil {
		cred = getCredentials(c.Value)
	}
	if cred == nil && !h.optional {
		http.Error(w, "Not logged in.", 403)
		return
	}
	h.handler(w, r, cred)
}

var signinOAuthClient oauth.Client

var (
	// secrets maps credential tokens to credential secrets. A real application will use a database to store credentials.
	secretsMutex sync.Mutex
	secrets      = map[string]string{}
)

func putCredentials(cred *oauth.Credentials) {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	secrets[cred.Token] = cred.Secret
}

func getCredentials(token string) *oauth.Credentials {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	if secret, ok := secrets[token]; ok {
		return &oauth.Credentials{Token: token, Secret: secret}
	}
	return nil
}

func deleteCredentials(token string) {
	secretsMutex.Lock()
	defer secretsMutex.Unlock()
	delete(secrets, token)
}

func serveSignin(w http.ResponseWriter, r *http.Request) {
	var (
		url string
		err error
	)

	service = &oauth1a.Service{
		RequestURL:   "https://api.twitter.com/oauth/request_token",
		AuthorizeURL: "https://api.twitter.com/oauth/authenticate",
		AccessURL:    "https://api.twitter.com/oauth/access_token",
		ClientConfig: &oauth1a.ClientConfig{
			ConsumerKey:    twitterConsumerKey,
			ConsumerSecret: twitterConsumerSecret,
			CallbackURL:    "http://" + r.Host + "/callback",
		},
		Signer: new(oauth1a.HmacSha1Signer),
	}
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	httpClient := urlfetch.Client(c)
	userConfig := &oauth1a.UserConfig{}
	if err = userConfig.GetRequestToken(service, httpClient); err != nil {
		log.Printf("Could not get request token: %v", err)
		http.Error(w, "Problem getting the request token", 500)
		return
	}
	if url, err = userConfig.GetAuthorizeURL(service); err != nil {
		log.Printf("Could not get authorization URL: %v", err)
		http.Error(w, "Problem getting the authorization URL", 500)
		return
	}
	log.Printf("Redirecting user to %v\n", url)
	http.Redirect(w, r, url, 302)
	var tempCred oauth.Credentials
	tempCred.Token = userConfig.RequestTokenKey
	putCredentials(&tempCred)

}

// serveOAuthCallback handles callbacks from the OAuth server.
func serveOAuthCallback(w http.ResponseWriter, r *http.Request) {
	tempCred := getCredentials(r.FormValue("oauth_token"))
	if tempCred == nil {
		http.Error(w, "Unknown oauth_token.", 500)
		return
	}
	deleteCredentials(tempCred.Token)

	var (
		err        error
		token      string
		verifier   string
		userConfig oauth1a.UserConfig
	)
	userConfig.RequestTokenKey = tempCred.Token
	log.Printf("Callback hit.\n")

	if token, verifier, err = userConfig.ParseAuthorize(r, service); err != nil {
		log.Printf("Could not parse authorization: %v", err)
		http.Error(w, "Problem parsing authorization", 500)
		return
	}
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	httpClient := urlfetch.Client(c)
	if err = userConfig.GetAccessToken(token, verifier, service, httpClient); err != nil {
		log.Printf("Error getting access token: %v", err)
		http.Error(w, "Problem getting an access token", 500)
		return
	}
	tokenCred := oauth.Credentials{Token: userConfig.AccessTokenKey,
		Secret: userConfig.AccessTokenSecret}
	if err != nil {
		http.Error(w, "Error getting request token, "+err.Error(), 500)
		return
	}
	putCredentials(&tokenCred)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Path:     "/",
		HttpOnly: true,
		Value:    tokenCred.Token,
	})

	http.Redirect(w, r, "/"+userConfig.AccessValues.Get("screen_name"), 302)
}

// serveLogout clears the authentication cookie.
func serveLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
	})
	http.Redirect(w, r, "/", 302)
}

func serveHome(w http.ResponseWriter, r *http.Request, cred *oauth.Credentials) {
	murl := strings.TrimLeft(r.URL.Path, "/@!#")
	var err error
	var tmpl *template.Template
	tmpl, err = template.ParseFiles("templates/base.html")
	tmpl.Execute(w, nil)
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	go_mobydick.SetConsumerKey(twitterConsumerKey)
	go_mobydick.SetConsumerSecret(twitterConsumerSecret)
	var api go_mobydick.TwitterApi
	if cred == nil {
		api = go_mobydick.NewTwitterApi(twitterAccessToken, twitterAccessTokenSecret)
	} else {
		api = go_mobydick.NewTwitterApi(cred.Token, cred.Secret)
	}

	if cred == nil {
		tmpl, err = template.ParseFiles("templates/navbar-nosigned.html")
		tmpl.Execute(w, nil)
	} else {
		search_result, err := api.GetVerifyCredentials()
		if err != nil {
			panic(err)
		}
		tmpl, err = template.ParseFiles("templates/navbar-signed.html")
		tmpl.Execute(w, search_result)
	}
	if err != nil {
		c.Errorf("%v", err)
	}
	if (murl == "/") || (murl == "") {
		tmpl, err = template.ParseFiles("templates/inicio.html", "templates/footer.html")
		if err != nil {
			c.Errorf("%v", err)
		}
		tmpl.Execute(w, nil)
		tmpl, err = template.ParseFiles("templates/footer.html")
		if err != nil {
			c.Errorf("%v", err)
		}
		tmpl.Execute(w, nil)
		return
	}

	v := url.Values{}
	v.Set("count", "1")
	v.Set("screen_name", murl)

	result, err := api.GetUserTimeline(v)
	if err != nil {
		c.Errorf("%v", err)
	}

	if len(result) == 0 {
		tmpl, err = template.ParseFiles("templates/no-tweet.html")
		if err != nil {
			c.Errorf("%v", err)
		}
		tmpl.Execute(w, nil)
	} else {

		//switch animo := analyze(result[0]); animo {
		switch animo := analyze2(result[0]); animo {
		//switch animo := valorar(result[0]); animo.Ponderacion {
		case "POSITIVA":
			tmpl, err = template.ParseFiles("templates/user-happy.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		case "NEGATIVA":
			tmpl, err = template.ParseFiles("templates/user-sad.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		default:
			tmpl, err = template.ParseFiles("templates/user.html")
			if err != nil {
				c.Errorf("%v", err)
			}
		}

		tmpl.Execute(w, result[0])
	}
	if cred != nil {
		tmpl, err = template.ParseFiles("templates/follows.html")
		if err != nil {
			c.Errorf("%v", err)
		}
		tmpl.Execute(w, murl)
	}
	tmpl, err = template.ParseFiles("templates/footer.html")
	if err != nil {
		c.Errorf("%v", err)
	}
	tmpl.Execute(w, nil)
	return
}

type tweetValorado struct {
	Screen_name       string
	Profile_image_url string
	Tweet             string
	Valoracion        string
}

type follow struct {
	Follows   []tweetValorado
	NextBatch string
}

func serveFollowing(w http.ResponseWriter, r *http.Request, cred *oauth.Credentials) {
	murl := strings.TrimPrefix(r.URL.Path, "/following/")
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	go_mobydick.SetConsumerKey(twitterConsumerKey)
	go_mobydick.SetConsumerSecret(twitterConsumerSecret)
	var api go_mobydick.TwitterApi
	api = go_mobydick.NewTwitterApi(cred.Token, cred.Secret)
	cads := strings.Split(murl, "/")
	v := url.Values{}
	v.Set("count", "10")
	v.Set("screen_name", cads[0])
	if len(cads) == 2 {
		v.Set("cursor", cads[1])
	}
	result, err := api.GetFriendsIds(v)
	v = url.Values{}
	var ids string
	for _, id := range result.Ids {
		ids += strconv.FormatInt(id, 10) + ", "
	}
	v.Set("user_id", ids)
	results, err := api.GetUsersLookup(v)
	if err != nil {
		c.Errorf("%v", err)
	}
	var data follow
	data.NextBatch = strconv.FormatInt(result.Next_cursor, 10)
	data.Follows = make([]tweetValorado, len(results))
	//fmt.Fprintf(w, "RECIBIDO:\n%#v", results)
	for i, user := range results {
		if user.Status != nil {
			data.Follows[i] = tweetValorado{*user.Screen_name, *user.Profile_image_url, user.Status.Text, string(analyze2(*user.Status))}
		} else {
			data.Follows[i] = tweetValorado{*user.Screen_name, *user.Profile_image_url, "El usuario introducido no existe o no tiene Tweets visibles.", "NEUTRA"}
		}
	}
	b, err := json.Marshal(data)
	fmt.Fprintf(w, "%s", b)
}

func serveFollowers(w http.ResponseWriter, r *http.Request, cred *oauth.Credentials) {
	murl := strings.TrimPrefix(r.URL.Path, "/followers/")
	c := appengine.NewContext(r)
	http.DefaultClient = urlfetch.Client(c)
	go_mobydick.SetConsumerKey(twitterConsumerKey)
	go_mobydick.SetConsumerSecret(twitterConsumerSecret)
	var api go_mobydick.TwitterApi
	api = go_mobydick.NewTwitterApi(cred.Token, cred.Secret)
	cads := strings.Split(murl, "/")
	v := url.Values{}
	v.Set("count", "10")
	v.Set("screen_name", cads[0])
	if len(cads) == 2 {
		v.Set("cursor", cads[1])
	}
	result, err := api.GetFollowersIds(v)
	v = url.Values{}
	var ids string
	for _, id := range result.Ids {
		ids += strconv.FormatInt(id, 10) + ", "
	}
	v.Set("user_id", ids)
	results, err := api.GetUsersLookup(v)
	if err != nil {
		c.Errorf("%v", err)
	}
	var data follow
	data.NextBatch = strconv.FormatInt(result.Next_cursor, 10)
	data.Follows = make([]tweetValorado, len(results))
	//fmt.Fprintf(w, "RECIBIDO:\n%#v", results)
	for i, user := range results {
		if user.Status != nil {
			data.Follows[i] = tweetValorado{*user.Screen_name, *user.Profile_image_url, user.Status.Text, string(analyze2(*user.Status))}
		} else {
			data.Follows[i] = tweetValorado{*user.Screen_name, *user.Profile_image_url, "El usuario introducido no existe o no tiene Tweets visibles.", "NEUTRA"}
		}
	}
	b, err := json.Marshal(data)
	fmt.Fprintf(w, "%s", b)
}
