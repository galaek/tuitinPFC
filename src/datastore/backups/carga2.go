package main

import (
	"time"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	//"strings"
	"appengine"
	"appengine/datastore"
	"appengine/remote_api"
	"fmt"
	"os/exec"
	"golang.org/x/oauth2"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
)

var (
	host                = flag.String("host", "manufirstapp.appspot.com", "hostname of application")
	email               = flag.String("email", "manugracia@gmail.com", "email of an admin user for the application")
	passwordFile        = flag.String("password", "pass2.txt", "file which contains the user's password")
)
const DatastoreKindName = "Greeting"


func newOAuthClient(ctx context.Context, config *oauth2.Config) *http.Client {
	token := tokenFromWeb(ctx, config)
	return config.Client(ctx, token)
}

func tokenFromWeb(ctx context.Context, config *oauth2.Config) *oauth2.Token {
	ch := make(chan string)
	randState := fmt.Sprintf("st%d", time.Now().UnixNano())
	ts := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/favicon.ico" {
			http.Error(rw, "", 404)
			return
		}
		if req.FormValue("state") != randState {
			log.Printf("State doesn't match: req = %#v", req)
			http.Error(rw, "", 500)
			return
		}
		if code := req.FormValue("code"); code != "" {
			fmt.Fprintf(rw, "<h1>Success</h1>Authorized.")
			rw.(http.Flusher).Flush()
			ch <- code
			return
		}
		log.Printf("no code")
		http.Error(rw, "", 500)
	}))
	defer ts.Close()

	config.RedirectURL = ts.URL
	config.RedirectURL = "http://localhost:8080/oauth2callback/"
	//config.RedirectURL = "http://manufirstapp.appspot.com/oauth2callback/"
	authURL := config.AuthCodeURL(randState)
	go openURL(authURL)
	log.Printf("Authorize this app at: %s", authURL)
	code := <-ch
	log.Printf("Got code: %s", code)

	token, err := config.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Token exchange error: %v", err)
	}
	return token
}

func openURL(url string) {
	try := []string{"xdg-open", "google-chrome", "open"}
	for _, bin := range try {
		err := exec.Command(bin, url).Run()
		if err == nil {
			return
		}
	}
	log.Printf("Error opening URL in browser.")
}

func main() {
	flag.Parse()

	conf := &oauth2.Config{ 
		//ClientID:     "a0b445ce731a4c41c3b8be2ab8a8a522291b95f3",
		ClientID:     "687230787099-6lq0raomm38vps5vbg6pmmp17lr4b0i0.apps.googleusercontent.com",
		ClientSecret: "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQD3gJOir0W/HHbx\nG2gVi7rT6cVPWNc71UN91h1YsdSwUba9Y92cWjVtRrDFaQX4f7EB0axlIG803oyd\n0ftL0jA+dWnU0Vbtz94yAVvexVgwSKOr0OQ6Gfm+uZ0k04A1caC6usn9znM3YMCa\nwCwDASC1sYvFU2p9bsOVCaba3ao12oMwGi0xWOsqFqmjnMv3+KlBj+ChQHGczjdh\nTvYbX5gAkxD9+9/YpbZId2+mF9SRwnPS21hiOG/NPDlCS/6reux+l/0E6+eRyojm\nu6d397dCHezefI4GYLtMKXFNY9lsRijVPlvl4x97e76mgZPj1d52TTkB7PTU1wIs\nJIpJrSzHAgMBAAECggEABZnAPX8v2dpACsau/UXTLXZtw5TkEfOKem987IPhpzfC\nJdj7q80Sxm1CFMWCoBPronnnJ8arHYwnrG6S/C0+cDth8LHoAKuigIktVgYrL7SF\nF587euEZmKpElw++J/dxRqhxZ6/jRY6H7TiKBmthHRtuaUGw+DOoc1frkapQbrV+\nvB1nU4y4vCFzZFJT36Krmjl1xGmEm72F2gWX5KpjzH5NV30Zur47iAubgJXs8r47\nS4RWioc6I6G7uuzzqFKa2HuYsmI3Gxzke09CI9HGPl1aG/vvexru8Sc93/KuWZXH\n7utWUDfFJGjIal8+MheTQJ8tX8jkKjb67WZhaJGycQKBgQD7yNUtWcectLikKokV\necCQXzYweAY7XO3C72PRrqGWVUyUs34VvWnlJf5btPyyQJjimkVrf80SauGZrLIy\nltO87lrVcB/JU0cCSCNk+Wdmfv8FcXUcehne5tb9tbwStq7YATx5eh2GnDi0s8P2\nA5zgd2QTPn5FuiYWZjg1vl50nwKBgQD7pWPSn3TyedLMB13oNwEpeWJ+S4e3U7St\n1Z4qxGLICHlwEiPQsWwan8X5JuP5opwCgVLD0sAuHB5QwSgcTZLJhy+otOufp6Q4\nu0uJmgaNKACc96M97TguCBLr4MBo7df1vCOBe8+4QjAUn8SYo6vmT/gqmnmRZREM\nKSDXOoBu2QKBgQDl0g8jcguNsjfHQTwXaiamoQGphCTMEqrDgBcw0aGUww8/vAae\neWIrU160/qKZYfUrAX3T/beF1CFQUB3np1xl23r1z350GZt7LbWA+VW0bL8CjOlE\nsP7kQviCZFvjCPTXHWnByAEjWX05E80OxYVwLgoetrAznRIe5/but3EoKQKBgQC9\nzY5QM9NKfFZha4EKAErRFGwUpDV2Mh2KLCBDU6LKC5JE1HnNE7VdE3uIJCw5gsu3\nHAHoD5LCdJTtBfOR/XSkqmFpFyTNY+16mNIttE4Ss8RaoHGw6LbCCXb0EK4vto14\nHHKPXGpdKRcIx0TKeFDUwyaEQ8VDw/4qtO6/R7HNaQKBgBKumCuHs+fUoYQJJEto\neTX3li6HTyw/Y5ijTvtdiRm4aRj8I8U5umIOWh3jHwF4hKDsFvd1SgfGMFkl6XgP\n0iR6xgiIO+GyKy2o3NR23pjSHnPhCctJY8AiQGAtYYi5UceZ+0vzo/3uJ4sayKU5\nqyk+fyOJswOqMpCMlQRJ+lCx\n-----END PRIVATE KEY-----\n",
		Scopes:       []string{"email"},
		Endpoint: google.Endpoint,
		/*Endpoint: oauth2.Endpoint{
			AuthURL:  "http://manufirstapp.appspot.com/_ah/OAuthAuthorizeToken", //"https://provider.com/o/oauth2/auth",
			TokenURL: "http://manufirstapp.appspot.com/_ah/OAuthGetRequestToken", //"https://provider.com/o/oauth2/token",
		},*/
	}
	//ctx := context.Background()
	ctx := oauth2.NoContext
	/**func (c *Config) PasswordCredentialsToken(ctx context.Context, username, password string) (*Token, error)*/
	client := newOAuthClient(ctx, conf)
	//client := clientLoginClient(*host, *email, password)
	c, err := remote_api.NewRemoteContext(*host, client)
	if err != nil {
		log.Fatalf("Failed to create context: %v", err)
	}
	log.Printf("App ID %q", appengine.AppID(c)) 


	
// q := datastore.NewQuery("surveyresult").KeysOnly()
// keys, err := q.GetAll(c, nil)
// fmt.Println(len(keys))


/*
q := datastore.NewQuery("surveyresult")
// Iterate over the results.
t := q.Run(c)
i:=0
for {
        var p SurveyResult
        _, err := t.Next(&p)
        if err == datastore.Done {
                break
        }
        if err != nil {
                c.Errorf("fetching next Person: %v", err)
                break
        }
		fmt.Println(i,p)
		i++
        // Do something with the Person p
}	
*/	
	var kinds []Greeting
	q := datastore.NewQuery("Greeting").Limit(50000)

	if _, err := q.GetAll(c, &kinds); err != nil {
		log.Fatalf("Failed to fetch kind info: %v", err)
	}
	for i:=0; i<len(kinds); i++ {
		fmt.Println(kinds[i])
	}
	fmt.Println(len(kinds))
}
type ProfileInfo struct {
	Nombre		string
	Descripción	string
	Hashtag		string
	ImagenURL	string
	ScreenName	string
	NombrePlano	string
}

type Greeting struct {
	Id			int
	Author		string
	Content		string
	Date		string
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

func clientLoginClient(host, email, password string) *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("failed to make cookie jar: %v", err)
	}
	client := &http.Client{
		Jar: jar,
	}

	v := url.Values{}
	v.Set("Email", email)
	v.Set("Passwd", password)
	v.Set("service", "ah")
	v.Set("source", "Misc-remote_api-0.1")
	v.Set("accountType", "HOSTED_OR_GOOGLE")

	resp, err := client.PostForm("https://www.google.com/accounts/ClientLogin", v)
	if err != nil {
		log.Fatalf("could not post login: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("unsuccessful request: status %d; body %q", resp.StatusCode, body)
	}
	if err != nil {
		log.Fatalf("unable to read response: %v", err)
	}

	m := regexp.MustCompile(`Auth=(\S+)`).FindSubmatch(body)
	if m == nil {
		log.Fatalf("no auth code in response %q", body)
	}
	auth := string(m[1])

	u := &url.URL{
		Scheme:   "https",
		Host:     host,
		Path:     "/_ah/login",
		RawQuery: "continue=/&auth=" + url.QueryEscape(auth),
	}

	// Disallow redirects.
	redirectErr := errors.New("stopping redirect")
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return redirectErr
	}

	resp, err = client.Get(u.String())
	if urlErr, ok := err.(*url.Error); !ok || urlErr.Err != redirectErr {
		log.Fatalf("could not get auth cookies: %v", err)
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusFound {
		log.Fatalf("unsuccessful request: status %d; body %q", resp.StatusCode, body)
	}

	client.CheckRedirect = nil
	return client
}
