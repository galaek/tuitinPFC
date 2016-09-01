package profiles

// Using application-only auth.

import (
	"encoding/json"
	"fmt"
	"github.com/kurrik/oauth1a"
	"github.com/kurrik/twittergo"
	//"github.com/araddon/httpstream"
	"twittertypes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"labix.org/v2/mgo"	
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"strconv"
	//"errors"

)

// Agregar dos variables globales tipo string para la consumerKey y la consumerSecret del usuario ke hace sign_in
// Modificar "LoadCredentials" poniendo un IF para que si las variables contienen algo, las use en el cliente

var (
		UserKey = ""
		UserSecret = ""
		AppKey = "Cij69U0URS7d0dMNkw9p2nfJp"
		AppSecret = "gl5oZunkPrOWcZqqFlmyMQcnlyhVEIGObGCVTJuh8cvTeaSI9Z"
	)
var (
		Sessions map[string]*oauth1a.UserConfig
	)
	
func LoadCredentials() (client *twittergo.Client, err error) {
	credentials, err := ioutil.ReadFile("CREDENTIALS")
	if err != nil {
		return
	}
	lines := strings.Split(string(credentials), "\n")
	config := &oauth1a.ClientConfig{
		ConsumerKey:    lines[0],
		ConsumerSecret: lines[1],
	}
	if (UserKey == "") || (UserSecret == "") {
		client = twittergo.NewClient(config, nil)
	} else {
		user := oauth1a.NewAuthorizedConfig(UserKey, UserSecret)
		client = twittergo.NewClient(config, user)
	}
	
	return
}

func LoadCredentialsUser(sessionID string) (client *twittergo.Client, err error) {
	// credentials, err := ioutil.ReadFile("CREDENTIALS")
	// if err != nil {
		// return
	// }
	// lines := strings.Split(string(credentials), "\n")
	config := &oauth1a.ClientConfig{
		ConsumerKey:    AppKey,
		ConsumerSecret: AppSecret,
	}	
	oauthToken := Sessions[sessionID].AccessValues.Get("oauth_token")
	oauthSecret := Sessions[sessionID].AccessValues.Get("oauth_token_secret")

	user := oauth1a.NewAuthorizedConfig(oauthToken, oauthSecret)
	client = twittergo.NewClient(config, user)	
	return
}

// GetMostRetweetedAndFavorited returns the 5 most retweeted tweets and the 5 most favorited tweets in the collection
func GetMostRetweetedAndFavorited (colec mgo.Collection) Popularity {
	var ids []struct{Id int64}
	err := colec.Find(nil).Select(bson.M{"_id":0, "id": 1}).All(&ids)
	if err != nil { 
		panic(err) 
	}
	fmt.Println("Numero de ids:", len(ids))
	var (
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		//query   url.Values
		//user 	twittergo.User
		tweet	[]twittergo.Tweet
		text	[]byte
	)
	// We have now the IDs list to make the Twitter Query
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		minwait     = time.Duration(10) * time.Second
		urltmpl		= "/1.1/statuses/lookup.json?"
	)
	endpoint := "/1.1/statuses/lookup.json?id="
	total := 0
	received := 0
	count := 0
	remaining := 100
	var tweets []twittertypes.Tweet	
	var top5 Popularity
	top5.MinRt = -1	
	top5.MinFav = -1	
	for total<len(ids) {
	//for total<150 {
		if remaining > 0 {
			// We add another ID to the query
			endpoint += strconv.FormatInt(ids[total].Id, 10) + ","
			count++
			total++
			remaining--
		} else {
			// We reset counters and make the query
			//fmt.Println("-->", endpoint, "<--")
				RETRY_IN:
				if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
					fmt.Printf("Could not parse request: %v\n", err)
					os.Exit(1)
				}
				if resp, err = client.SendRequest(req); err != nil {
					fmt.Println(err)
					fmt.Printf("Could not send request: %v\n", err)
					os.Exit(1)
				}
				defer resp.Body.Close()
				//user = twittergo.User{}
				tweet = []twittergo.Tweet{}
				if err = resp.Parse(&tweet); err != nil {
					if rle, ok := err.(twittergo.RateLimitError); ok {
						dur := rle.Reset.Sub(time.Now()) + time.Second
						if dur < minwait {
							// Don't wait less than minwait.
							dur = minwait
						}
						msg := "Rate limited. Reset at %v. Waiting for %v\n"
						fmt.Printf(msg, rle.Reset, dur)
						time.Sleep(dur)
						goto RETRY_IN // Retry request.
					} else {
						fmt.Printf("Problem parsing response: %v\n", err)
					}
				}
				if resp.HasRateLimit() {
							fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
				}
				if text, err = json.Marshal(tweet); err != nil {
						fmt.Printf("Could not encode User: %v\n", err)
						os.Exit(1)
				}
				// We get the users and we store them in the DB
				json.Unmarshal(text, &tweets)
				for i:=0; i<len(tweets); i++ {
					if (AlreadyInRtList(tweets[i], top5) == false) && (tweets[i].Retweet_count > top5.MinRt) {
						// Add tweet to the list and update the rest if neccessary
						UpdateRts(tweets[i], &top5)
					}
					if (AlreadyInFavList(tweets[i], top5) == false) && tweets[i].Favorite_count > top5.MinFav {
						// Add tweet to the list and update the rest if neccessary
						UpdateFavs(tweets[i], &top5)
					}
				}
				CalcMaxs(&top5)
				//fmt.Println(len(tweets), total)
			received = received + len(tweets)	
			count = 0
			remaining = 100
			endpoint = "/1.1/statuses/lookup.json?id="
		}
	}
	//fmt.Println("-->", endpoint, "<--")
	RETRY_OUT:
	if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
		fmt.Printf("Could not parse request: %v\n", err)
		os.Exit(1)
	}
	if resp, err = client.SendRequest(req); err != nil {
		fmt.Println(err)
		fmt.Printf("Could not send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	//user = twittergo.User{}
	tweet = []twittergo.Tweet{}
	if err = resp.Parse(&tweet); err != nil {
		if rle, ok := err.(twittergo.RateLimitError); ok {
			dur := rle.Reset.Sub(time.Now()) + time.Second
			if dur < minwait {
				// Don't wait less than minwait.
				dur = minwait
			}
			msg := "Rate limited. Reset at %v. Waiting for %v\n"
			fmt.Printf(msg, rle.Reset, dur)
			time.Sleep(dur)
			goto RETRY_OUT // Retry request.
		} else {
			fmt.Printf("Problem parsing response: %v\n", err)
		}
	}
	if resp.HasRateLimit() {
				fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
	}
	if text, err = json.Marshal(tweet); err != nil {
			fmt.Printf("Could not encode User: %v\n", err)
			os.Exit(1)
	}
	// We get the users and we store them in the DB
	json.Unmarshal(text, &tweets)	
	for i:=0; i<len(tweets); i++ {
		if (AlreadyInRtList(tweets[i], top5) == false) && (tweets[i].Retweet_count > top5.MinRt) {
			// Add tweet to the list and update the rest if neccessary
			UpdateRts(tweets[i], &top5)
		}
		if (AlreadyInFavList(tweets[i], top5) == false) && tweets[i].Favorite_count > top5.MinFav {
			// Add tweet to the list and update the rest if neccessary
			UpdateFavs(tweets[i], &top5)
		}
	}
	CalcMaxs(&top5)
	//fmt.Println(len(tweets), total)
	top5.MostRt = OrderTopFiveRt(top5.MostRt)
	top5.MostFav = OrderTopFiveFav(top5.MostFav)
	// ShowTime
	received = received + len(tweets)
	top5.TotalTweets = received
	return top5
}

func OrderTopFiveRt (top [5]twittertypes.Tweet) [5]twittertypes.Tweet{
	var orderedTop5 [5]twittertypes.Tweet
	var demFive [5]twittertypes.Tweet
	demFive = top
	for j:=0; j<5; j++ {
		var maxValue int = 0
		var maxIndex int = -1
		for i:=0; i<5; i++ {
			if demFive[i].Retweet_count > maxValue {
				maxValue = demFive[i].Retweet_count
				maxIndex = i
			}
		}
		orderedTop5[j] = demFive[maxIndex]
		demFive[maxIndex].Retweet_count = 0
	}
	return orderedTop5
}

func OrderTopFiveFav (top [5]twittertypes.Tweet) [5]twittertypes.Tweet{
	var orderedTop5 [5]twittertypes.Tweet
	var demFive [5]twittertypes.Tweet
	demFive = top
	for j:=0; j<5; j++ {
		var maxValue int = 0
		var maxIndex int = -1
		for i:=0; i<5; i++ {
			if demFive[i].Favorite_count > maxValue {
				maxValue = demFive[i].Favorite_count
				maxIndex = i
			}
		}
		orderedTop5[j] = demFive[maxIndex]
		demFive[maxIndex].Favorite_count = 0
	}
	return orderedTop5
}

func AlreadyInRtList (tweet twittertypes.Tweet, top Popularity) bool {
	for i:=0; i<len(top.MostRt); i++ {
		if 	((top.MostRt[i].Id == tweet.Id) || (top.MostRt[i].Retweeted_status.Id == tweet.Retweeted_status.Id)) ||
			((top.MostRt[i].Id == tweet.Retweeted_status.Id) || (top.MostRt[i].Retweeted_status.Id == tweet.Id)){
			return true
		}
	}
	return false
}

func AlreadyInFavList (tweet twittertypes.Tweet, top Popularity) bool {
	for i:=0; i<len(top.MostFav); i++ {
		if (top.MostFav[i].Id == tweet.Id) {
			return true
		}
	}
	return false
}

func UpdateRts(tweet twittertypes.Tweet, top *Popularity) {
	var lowerRt int = 0
	for i:=0; i<len(top.MostRt); i++ {
		if top.MostRt[i].Retweet_count < top.MostRt[lowerRt].Retweet_count {
			lowerRt = i
		}
	}
	top.MostRt[lowerRt] = tweet
	CalcMins(top)
}

func UpdateFavs(tweet twittertypes.Tweet, top *Popularity) {
	var lowerFav int = 0
	for i:=0; i<len(top.MostFav); i++ {
		if top.MostFav[i].Favorite_count < top.MostFav[lowerFav].Favorite_count {
			lowerFav = i
		}
	}
	top.MostFav[lowerFav] = tweet
	CalcMins(top)	
}

func CalcMaxs (top *Popularity) {
	top.MaxRt = top.MostRt[0].Retweet_count
	top.MaxFav = top.MostFav[0].Favorite_count
	for i:=0; i<len(top.MostRt); i++ {
		if top.MostRt[i].Retweet_count > top.MaxRt {
			top.MaxRt = top.MostRt[i].Retweet_count
		}
		if top.MostFav[i].Favorite_count > top.MaxFav {
			top.MaxFav = top.MostFav[i].Favorite_count
		}
	}
}

func CalcMins (top *Popularity) {
	top.MinRt = top.MostRt[0].Retweet_count
	top.MinFav = top.MostFav[0].Favorite_count
	for i:=0; i<len(top.MostRt); i++ {
		if top.MostRt[i].Retweet_count < top.MinRt {
			top.MinRt = top.MostRt[i].Retweet_count
		}
		if top.MostFav[i].Favorite_count < top.MinFav {
			top.MinFav = top.MostFav[i].Favorite_count
		}
	}
}

type Popularity struct {
	// Id			int64
	MinRt			int
	MinFav			int
	MaxRt			int
	MaxFav			int
	MostRt			[5]twittertypes.Tweet
	MostFav			[5]twittertypes.Tweet
	TotalTweets		int
}

// GetCountryAndStateByCoordinates returns the country and state/province of the [latitud,longitud] given
func GetCountryAndStateByCoordinates(lat float64, long float64) (CountryAndState, error) {
	type countryResponse map[string]interface{}
	var response countryResponse
	var place CountryAndState
	var Url *url.URL
    Url, err := url.Parse("http://api.geonames.org/countrySubdivisionJSON")
    if err != nil {
        panic("Invalid URL")
    }
    parameters := url.Values{}
    parameters.Add("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	parameters.Add("lng", strconv.FormatFloat(long, 'f', 6, 64))
	parameters.Add("lang", "es") // This parameter works mainly for Spain
    parameters.Add("username", "manupruebas")
    Url.RawQuery = parameters.Encode()
	
	resp, err := http.Get(Url.String())
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()	
	body, err := ioutil.ReadAll(resp.Body)
	//fmt.Println(string(body))
	json.Unmarshal(body, &response)
	//fmt.Println(response)
	if response["status"] == nil {
		place.Country = response["countryName"].(string)
		place.State = response["adminName1"].(string)
	} else {
		err = &APIerror{int(response["status"].(map[string]interface{})["value"].(float64)), response["status"].(map[string]interface{})["message"].(string)}
	}
	return place, err
}

 	

func (e *APIerror) Error() string {
    return fmt.Sprintf("GeonamesAPI error %d - %s", e.Value, e.Message)
}

type APIerror struct {
	Value		int
	Message		string
}

type CountryAndState struct {
	Country 		string
	State 			string
	// CountryName			string
	// AdminName1			string
}

// GetFollowersByScreenNameToDB gets the 'followers' from the 'screenName' user and stores it in the DB
func GetFollowersByScreenNameToDB (c mgo.Collection, screenName string) int {
		var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	twittergo.User
		text	[]byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/followers/list.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	var total int = 0;
	var cursors Cursors	
	cursors.Next_cursor = -1
	for {
		if cursors.Next_cursor == 0 {
			break
		}
		query = url.Values{}
		query.Set("screen_name", screenName)
		query.Set("cursor", strconv.FormatInt(cursors.Next_cursor,10))
		query.Set("count", "200")
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
		}

		if text, err = json.Marshal(user); err != nil {
			fmt.Printf("Could not encode User: %v\n", err)
			os.Exit(1)
		}
		
		// We get the new cursors
		json.Unmarshal(text, &cursors)

		// We get the users and we store them in the DB
		var respuesta map[string]interface{}
		json.Unmarshal(text, &respuesta)
		
		var user1 twittertypes.UserProfile
		for i:=0; i<len(respuesta["users"].([]interface{})); i++ {
			if text, err = json.Marshal(respuesta["users"].([]interface{})[i].(map[string]interface{})); err != nil {
				fmt.Printf("Could not encode User: %v\n", err)
				os.Exit(1)
			}
			json.Unmarshal(text, &user1)
			// Guardamos user1 en la BD
			err = c.Insert(user1) 
			if err != nil { 
				panic(err) 
			}
			total++
		}
	}
	return total
}

// GetFollowersByIDToDB gets the 'followers' from the 'id' user and stores it in the DB
func GetFollowersByIDToDB (c mgo.Collection, id int64) int {
		var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	twittergo.User
		text	[]byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/followers/list.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	var total int = 0;
	var cursors Cursors	
	cursors.Next_cursor = -1
	for {
		if cursors.Next_cursor == 0 {
			break
		}
		query = url.Values{}
		query.Set("user_id", strconv.FormatInt(id,10))
		query.Set("cursor", strconv.FormatInt(cursors.Next_cursor,10))
		query.Set("count", "200")
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
		}

		if text, err = json.Marshal(user); err != nil {
			fmt.Printf("Could not encode User: %v\n", err)
			os.Exit(1)
		}
		
		// We get the new cursors
		json.Unmarshal(text, &cursors)

		// We get the users and we store them in the DB
		var respuesta map[string]interface{}
		json.Unmarshal(text, &respuesta)
		
		var user1 twittertypes.UserProfile
		for i:=0; i<len(respuesta["users"].([]interface{})); i++ {
			if text, err = json.Marshal(respuesta["users"].([]interface{})[i].(map[string]interface{})); err != nil {
				fmt.Printf("Could not encode User: %v\n", err)
				os.Exit(1)
			}
			json.Unmarshal(text, &user1)
			// Guardamos user1 en la BD
			err = c.Insert(user1) 
			if err != nil { 
				panic(err) 
			}
			total++
		}
	}
	return total
}

// GetFriendsByIDToDB gets the 'friends' from the 'id' user and stores it in the DB
func GetFriendsByIDToDB (c mgo.Collection, id int64) int {
		var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	twittergo.User
		text	[]byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/friends/list.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	var total int = 0;
	var cursors Cursors	
	cursors.Next_cursor = -1
	for {
		if cursors.Next_cursor == 0 {
			break
		}
		query = url.Values{}
		query.Set("user_id", strconv.FormatInt(id,10))
		query.Set("cursor", strconv.FormatInt(cursors.Next_cursor,10))
		query.Set("count", "200")
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
		}

		if text, err = json.Marshal(user); err != nil {
			fmt.Printf("Could not encode User: %v\n", err)
			os.Exit(1)
		}
		
		// We get the new cursors
		json.Unmarshal(text, &cursors)

		// We get the users and we store them in the DB
		var respuesta map[string]interface{}
		json.Unmarshal(text, &respuesta)
		
		var user1 twittertypes.UserProfile
		for i:=0; i<len(respuesta["users"].([]interface{})); i++ {
			if text, err = json.Marshal(respuesta["users"].([]interface{})[i].(map[string]interface{})); err != nil {
				fmt.Printf("Could not encode User: %v\n", err)
				os.Exit(1)
			}
			json.Unmarshal(text, &user1)
			// Guardamos user1 en la BD
			err = c.Insert(user1) 
			if err != nil { 
				panic(err) 
			}
			total++
		}
	}
	return total
}

// GetFriendsByScreenNameToDB gets the 'friends' from the 'screenName' user and stores it in the DB
func GetFriendsByScreenNameToDB (c mgo.Collection, screenName string) int {
		var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	twittergo.User
		text	[]byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/friends/list.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	var total int = 0;
	var cursors Cursors	
	cursors.Next_cursor = -1
	for {
		if cursors.Next_cursor == 0 {
			break
		}
		query = url.Values{}
		query.Set("screen_name", screenName)
		query.Set("cursor", strconv.FormatInt(cursors.Next_cursor,10))
		query.Set("count", "200")
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
		}

		if text, err = json.Marshal(user); err != nil {
			fmt.Printf("Could not encode User: %v\n", err)
			os.Exit(1)
		}
		
		// We get the new cursors
		json.Unmarshal(text, &cursors)

		// We get the users and we store them in the DB
		var respuesta map[string]interface{}
		json.Unmarshal(text, &respuesta)
		
		var user1 twittertypes.UserProfile
		for i:=0; i<len(respuesta["users"].([]interface{})); i++ {
			if text, err = json.Marshal(respuesta["users"].([]interface{})[i].(map[string]interface{})); err != nil {
				fmt.Printf("Could not encode User: %v\n", err)
				os.Exit(1)
			}
			json.Unmarshal(text, &user1)
			// Guardamos user1 en la BD
			err = c.Insert(user1) 
			if err != nil { 
				panic(err) 
			}
			total++
		}
	}
	return total
}

type Cursors struct {
	Previous_cursor			int64
	Next_cursor				int64
}

// GetUserByScreenName gets the Twitter Profile corresponding to the given 'screenName'
func GetUserByScreenName (screenName string) twittertypes.UserProfile {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	[]twittergo.User
		text    []byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/users/lookup.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	
	query = url.Values{}
	query.Set("screen_name", screenName)
	endpoint := fmt.Sprintf(urltmpl, query.Encode())
	done := false
	for !done {
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = []twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
					dur := rle.Reset.Sub(time.Now()) + time.Second
					if dur < minwait {
						// Don't wait less than minwait.
						dur = minwait
					}
					msg := "Rate limited. Reset at %v. Waiting for %v\n"
					fmt.Printf(msg, rle.Reset, dur)
					time.Sleep(dur)
					// Retry request.
				} else {
					fmt.Printf("Problem parsing response: %v\n", err)
				}
		} else {
			done = true
		}
	}
	if resp.HasRateLimit() {
				fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
	}
	
	user1 := user[0]
	if text, err = json.Marshal(user1); err != nil {
		fmt.Printf("Could not encode User: %v\n", err)
		os.Exit(1)
	}
	var myuser twittertypes.UserProfile
	json.Unmarshal(text, &myuser)
	return myuser
}

// GetUserByID gets the Twitter Profile corresponding to the given 'id'
func GetUserByID (id int64) twittertypes.UserProfile {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		query   url.Values
		user 	[]twittergo.User
		text    []byte
	)
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		urltmpl     = "/1.1/users/lookup.json?%v"
		minwait     = time.Duration(10) * time.Second
	)
	query = url.Values{}
	query.Set("user_id", strconv.FormatInt(id, 10))
	endpoint := fmt.Sprintf(urltmpl, query.Encode())
	done := false
	for !done {
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()
		user = []twittergo.User{}
		if err = resp.Parse(&user); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
					dur := rle.Reset.Sub(time.Now()) + time.Second
					if dur < minwait {
						// Don't wait less than minwait.
						dur = minwait
					}
					msg := "Rate limited. Reset at %v. Waiting for %v\n"
					fmt.Printf(msg, rle.Reset, dur)
					time.Sleep(dur)
					// Retry request.
				} else {
					fmt.Printf("Problem parsing response: %v\n", err)
				}
		} else {
			done = true
		}	
	}
	if resp.HasRateLimit() {
				fmt.Printf("%v calls available (%s)\n", resp.RateLimitRemaining(), urltmpl)
	}
	
	user1 := user[0]
	if text, err = json.Marshal(user1); err != nil {
		fmt.Printf("Could not encode User: %v\n", err)
		os.Exit(1)
	}
	var myuser twittertypes.UserProfile
	json.Unmarshal(text, &myuser)
	return myuser
}

// GetTimelineToDBcreds gets the 'maxTweets' last Tweets from a Twitter Timeline
func GetTimelineToDBcreds (colec mgo.Collection, screenName string, maxTweets int, sessionID string) {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		max_id  uint64
		//out     *os.File
		query   url.Values
		results *twittergo.Timeline
		text    []byte
	)
	var max int = maxTweets
	if maxTweets == 0 {
		max = -1 // We get the whole timeline
	}
	if client, err = LoadCredentialsUser(sessionID); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		count   int = 200
		urltmpl     = "/1.1/statuses/user_timeline.json?%v"
		minwait     = time.Duration(10) * time.Second
		maxresuls 	= 1000
	)
	query = url.Values{}
	query.Set("count", fmt.Sprintf("%v", count))
	query.Set("screen_name", screenName)
	total := 0
	for {
		if max_id != 0 {
			query.Set("max_id", fmt.Sprintf("%v", max_id))
		}
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		results = &twittergo.Timeline{}
		if err = resp.Parse(results); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		batch := len(*results)
		if batch == 0 {
			fmt.Printf("No more results, end of timeline.\n")
			break
		}
		
		for _, tweet := range *results {
			if text, err = json.Marshal(tweet); err != nil {
				fmt.Printf("Could not encode Tweet: %v\n", err)
				os.Exit(1)
			}
			//out.Write(text)
			//out.Write([]byte("\n"))
			max_id = tweet.Id() - 1
			total += 1
			var mytweet twittertypes.Tweet
			//fmt.Println(string(text))
			json.Unmarshal(text, &mytweet)
			tweetstamp, err := time.Parse(time.RubyDate, mytweet.Created_at)
			mytweet.Epoch = tweetstamp.Unix()
			err = colec.Insert(mytweet) 
			if err != nil { 
				panic(err) 
			}
			if (total == max) {
				fmt.Println(total, "tweets already obtained")
				if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)", resp.RateLimitRemaining(),urltmpl)
				}
				fmt.Printf(".\n")
				fmt.Printf("--------------------------------------------------------\n")
				fmt.Printf("Wrote %v Tweets to DB\n", total)
				return
			}
		}
		fmt.Printf("Got %v Tweets", batch)
		if resp.HasRateLimit() {
			fmt.Printf(", %v calls available (%s)", resp.RateLimitRemaining(), urltmpl)
		}
		fmt.Printf(".\n")
	}
	fmt.Printf("--------------------------------------------------------\n")
	fmt.Printf("Wrote %v Tweets to DB\n", total)
}

// GetTimelineToDB gets the 'maxTweets' last Tweets from a Twitter Timeline
func GetTimelineToDB (colec mgo.Collection, screenName string, maxTweets int) {
	var (
		err     error
		client  *twittergo.Client
		req     *http.Request
		resp    *twittergo.APIResponse
		max_id  uint64
		//out     *os.File
		query   url.Values
		results *twittergo.Timeline
		text    []byte
	)
	var max int = maxTweets
	if maxTweets == 0 {
		max = -1 // We get the whole timeline
	}
	if client, err = LoadCredentials(); err != nil {
		fmt.Printf("Could not parse CREDENTIALS file: %v\n", err)
		os.Exit(1)
	}
	const (
		count   int = 200
		urltmpl     = "/1.1/statuses/user_timeline.json?%v"
		minwait     = time.Duration(10) * time.Second
		maxresuls 	= 1000
	)
	query = url.Values{}
	query.Set("count", fmt.Sprintf("%v", count))
	query.Set("screen_name", screenName)
	total := 0
	for {
		if max_id != 0 {
			query.Set("max_id", fmt.Sprintf("%v", max_id))
		}
		endpoint := fmt.Sprintf(urltmpl, query.Encode())
		if req, err = http.NewRequest("GET", endpoint, nil); err != nil {
			fmt.Printf("Could not parse request: %v\n", err)
			os.Exit(1)
		}
		if resp, err = client.SendRequest(req); err != nil {
			fmt.Println(err)
			fmt.Printf("Could not send request: %v\n", err)
			os.Exit(1)
		}
		results = &twittergo.Timeline{}
		if err = resp.Parse(results); err != nil {
			if rle, ok := err.(twittergo.RateLimitError); ok {
				dur := rle.Reset.Sub(time.Now()) + time.Second
				if dur < minwait {
					// Don't wait less than minwait.
					dur = minwait
				}
				msg := "Rate limited. Reset at %v. Waiting for %v\n"
				fmt.Printf(msg, rle.Reset, dur)
				time.Sleep(dur)
				continue // Retry request.
			} else {
				fmt.Printf("Problem parsing response: %v\n", err)
			}
		}
		batch := len(*results)
		if batch == 0 {
			fmt.Printf("No more results, end of timeline.\n")
			break
		}
		
		for _, tweet := range *results {
			if text, err = json.Marshal(tweet); err != nil {
				fmt.Printf("Could not encode Tweet: %v\n", err)
				os.Exit(1)
			}
			//out.Write(text)
			//out.Write([]byte("\n"))
			max_id = tweet.Id() - 1
			total += 1
			var mytweet twittertypes.Tweet
			//fmt.Println(string(text))
			json.Unmarshal(text, &mytweet)
			tweetstamp, err := time.Parse(time.RubyDate, mytweet.Created_at)
			mytweet.Epoch = tweetstamp.Unix()
			err = colec.Insert(mytweet) 
			if err != nil { 
				panic(err) 
			}
			if (total == max) {
				fmt.Println(total, "tweets already obtained")
				if resp.HasRateLimit() {
					fmt.Printf("%v calls available (%s)", resp.RateLimitRemaining(),urltmpl)
				}
				fmt.Printf(".\n")
				fmt.Printf("--------------------------------------------------------\n")
				fmt.Printf("Wrote %v Tweets to DB\n", total)
				return
			}
		}
		fmt.Printf("Got %v Tweets", batch)
		if resp.HasRateLimit() {
			fmt.Printf(", %v calls available (%s)", resp.RateLimitRemaining(), urltmpl)
		}
		fmt.Printf(".\n")
	}
	fmt.Printf("--------------------------------------------------------\n")
	fmt.Printf("Wrote %v Tweets to DB\n", total)
}

// GetTweetsPerDay calculates the number of tweets per day, the average, the total days and the total tweets obtained
func GetTweetsPerDay (colec mgo.Collection) TweetsFrequency {
	var tweets []twittertypes.Tweet
	//err := colec.Find(bson.M{"text": bson.M{"$exists": true}}).All(&tweets)
	//iter := colec.Find(nil).Sort(bson.D{"epoch", 1}).Iter()
	//iter := colec.Find(nil).Sort(bson.D{{Name:"epoch", Value: 1}}).Iter()
	err := colec.Find(nil).Sort("-epoch").All(&tweets)
	if err != nil { 
		panic(err) 
	}
	var daysList []DayAndTweets
	var freq TweetsFrequency
	numDays := 0
	lastDayChecked := time.Now().AddDate(1,0,0) 
	for i:=0; i<len(tweets); i++ {
		tweetstamp, err := time.Parse(time.RubyDate, tweets[i].Created_at)
		if err != nil { 
			panic(err) 
		}
		if (lastDayChecked.Year() == tweetstamp.Year()) && (lastDayChecked.YearDay() == tweetstamp.YearDay()) {
			// Same day
			daysList[numDays-1].NumTweets = daysList[numDays-1].NumTweets + 1
		} else {
			// Different day
			lastDayChecked = tweetstamp
			numDays++
			var oneDay DayAndTweets
			oneDay.Day = tweetstamp
			oneDay.NumTweets = 1
			daysList = append(daysList, oneDay)
		}
	}
	
	freq.List = daysList
	tweet0, err := time.Parse(time.RubyDate, tweets[0].Created_at)
	tweetN, err := time.Parse(time.RubyDate, tweets[len(tweets)-1].Created_at)
	tweet0 = tweet0.AddDate(0,0,1) // Checking until 00:00h next day
	fmt.Println(tweets[0].Created_at)
	fmt.Println(tweets[len(tweets)-1].Created_at)
	
	first := time.Date(tweet0.Year(), tweet0.Month(), tweet0.Day(), 0, 0, 0, 0, time.UTC)
	last := time.Date(tweetN.Year(), tweetN.Month(), tweetN.Day(), 0, 0, 0, 0, time.UTC)
	freq.TotalDays = DaysFrom(last, first)
	freq.TotalTweets = (len(tweets))
	freq.Average = float64(freq.TotalTweets) / float64(freq.TotalDays)
	return freq
}

// GetLocation returns the location provided by the user in the profile
func GetLocation (user twittertypes.UserProfile) string {
	return user.Location
}

func DetermineGender (name string) (string, int){
	var gen Gender
	gen, err := GenderFromName(name)
	if err != nil {
		fmt.Println(err)
	} else {
		if gen.Gender == "null" {
			//fmt.Println("Complete name not found")
			names := strings.Split(name, " ")
			for i:=0; i<len(names); i++ {
				gen, err = GenderFromName(names[i])
				if err != nil {
					fmt.Println(err)
					break
				} else {
					if (gen.Gender != "null") && (gen.Probability > 0.5) {
						//fmt.Println("Found match on part:", i)
						break
					}
				}
			}
		} else {
			//fmt.Println("Complete name found")
		}
		if gen.Gender == "null" {
			//fmt.Println("Unable to determine gender")
		} else {
			//fmt.Println(gen.Probability)
			return gen.Gender, int(gen.Probability * 100) //+ "(pretty sure)"			
		}	
	}
	return "Not determined", 0
}

// GenderFromName tries to determine the gender from the person named 'name'
func GenderFromName(name string) (Gender, error) {
	type genderResponse map[string]interface{}
	var err error
	var gen genderResponse
	var gender Gender
	var Url *url.URL
    Url, err = url.Parse("http://api.genderize.io")
    if err != nil {
        panic("Invalid URL")
    }
    parameters := url.Values{}
    parameters.Add("name", name)
    parameters.Add("country_id", "es")
    Url.RawQuery = parameters.Encode()
	
	resp, err := http.Get(Url.String())
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()	

	body, err := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &gen)
	
	if resp.Status != "200 OK" {
		// Error in the API Query
		err = &GenderAPIerror{resp.Status, gen["error"].(string)}
		return gender, err
	}
	if gen["gender"] == nil { 
		// We dont use the spanish country code
		Url, err := url.Parse("http://api.genderize.io")
		if err != nil {
			panic("Invalid URL")
		}
		parameters := url.Values{}
		parameters.Add("name", name)
		Url.RawQuery = parameters.Encode()
		resp, err := http.Get(Url.String())
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()	
		body, err := ioutil.ReadAll(resp.Body)
		json.Unmarshal(body, &gen)
		if resp.Status != "200 OK" {
			// Error in the API Query
			err = &GenderAPIerror{resp.Status, gen["error"].(string)}
			return gender, err
		}
	}
	
	if gen["gender"] == nil {
		gender.Name = gen["name"].(string) 
		gender.Gender = "null"
		gender.Probability = -1
		gender.Count = -1
	} else {
		gender.Name = gen["name"].(string)
		gender.Gender = gen["gender"].(string)
		gender.Probability, _ = strconv.ParseFloat(gen["probability"].(string), 64)
		gender.Count = int(gen["count"].(float64))
	}
	return gender, err
}

func (e *GenderAPIerror) Error() string {
    return fmt.Sprintf("GenderizeAPI error %s - %s", e.Status, e.Message)
}

type GenderAPIerror struct {
	Status		string
	Message		string
}

type Gender struct {
	Name			string
	Gender			string
	Probability		float64
	Count			int
}

type TweetsFrequency struct {
	Average 			float64
	List				[]DayAndTweets
	TotalDays			int
	TotalTweets			int
}

type DayAndTweets struct {
	Day					time.Time
	NumTweets			int
}

// DaysFrom returns the days from d1 to d2
func DaysFrom (d1, d2 time.Time) int {
	delta := d2.Sub(d1)
	return int(delta.Hours() / 24)
}

type WordWeight struct {
	Word		string
	Weight		int
}

//------------------------
// To Add a new Topic:
//------------------------
// 1) Create the map "Topics_XXXX" where XXXX is the name of the topic and containing the words for the topic
// 2) Increment in 1 the NumTopics const
// 3) Add the "TOpics_XXXX" word to the TopicsSlice variable in the LoadTopics function
// 4) Add the "XXXX" word in the TopicsList variable in the LoadTopics Function
//------------------------


// NumTopics is the number of topics we have stored
const NumTopics int = 11
// TopicsDB is a map of maps containing the words for the different topics and their weights
var TopicsDB map[string]map[string]int
// TopicsList contains a list of the name of the different topics in the TopicsDB
var TopicsList [NumTopics]string

// LoadTopics fills the TopicsDB map with the data.
func LoadTopics () {

	TopicsSlice := []map[string]int{Topics_football, Topics_basket, Topics_tenis, Topics_racing,
									Topics_sports, Topics_esports, Topics_music, Topics_technology, Topics_animals,
									Topics_arts, Topics_cinema}
	TopicsList = [NumTopics]string{"football", "basket", "tenis", "racing", "sports", "esports", "music",
							 "technology", "animals", "arts", "cinema"}
	TopicsDB = make(map[string]map[string]int)
	for i:=0; i<len(TopicsSlice); i++ {
		TopicsDB[TopicsList[i]] = TopicsSlice[i]
	}
}

var Topics_football = map[string]int{
		"fútbol": 1, "futbol": 1, "gol": 1, "portería": 1, "portero": 1,
		// Players
		"messi": 1, "cristiano": 1, "cristiano ronaldo": 1, "cr7": 1, 
		"ribery": 1, "ribéry": 1, 
		// Spain Players
		"xavi": 1, "ramos": 1, "iniesta": 1, "casillas": 1, "pepe reina": 1, 
		"arbeloa": 1, "albiol": 1, "sergio ramos": 1, "xabi alonso": 1,
		"cesc": 1, "silva": 1, "cazorla": 1, "torres": 1, "villa": 1,
		"diego costa": 1, "pique": 1, "jordi alba": 1, "busquets": 1,
		"pedro": 1, "juanfran": 1, "koke": 1, "de gea": 1, "azpilicueta": 1, 
		"mata": 1, "javi martinez": 1, "del bosque": 1, 
		// The rest of the players
}

var Topics_basket = map[string]int{
		"baloncesto": 1, "canasta": 1, "acb": 1, "nba": 1, "euroliga": 1,
		// Players
		"lebron": 1, "lebron james": 1, "gasol": 1, "ricky": 1, "rubio": 1,
		"ricky rubio": 1, "calderon": 1, "sergio rodriguez": 1, "ibaka": 1,
		"navarro": 1, "kevin durant": 1, "durant": 1, "ginobili": 1, 
		"tony parker": 1, "llull": 1, "felipe reyes": 1, "bryant": 1,	
}

var Topics_tenis =  map[string]int{
		// Players
		"nadal": 1, "djokovic": 1, "nole": 1, "wawrinka": 1, "ferrer": 1,
		"federer": 1, "murray": 1, "berdych": 1, "del potro": 1,
		"raonic": 1, "gulbis": 1, "nishikori": 1, "isner": 1, "dimitrov": 1,
		"gasquet": 1, "fognini": 1, "youzhny": 1, "tsonga": 1, "monfils": 1,
		"robredo": 1, "verdasco": 1, "almagro": 1, "cilic": 1, "feliciano": 1,
		"serena williams": 1, "halep": 1, "radwanska": 1, "sharapova": 1,
		"kvitova": 1, "jankovic": 1, "azarenka": 1, "ivanovic": 1, "errani": 1,
		"carla suarez": 1, "stosur": 1, "wozniacki": 1, "kuznetsova": 1,
		"stepanek": 1, "granollers": 1, "marc lopez": 1, "su-wei": 1,
		"anabel medina": 1, 
}

var Topics_racing =  map[string]int{
		// F1
		"f1": 1, "formula 1": 1, 
		"fernando alonso": 1, "vettel": 1, "ricciardo": 1, "rosberg": 1,
		"hamilton": 1, "jenson button": 1, "hulkenberg": 1, "bottas": 1, 
		"vergne": 1, "magnussen": 1, "raikkonen": 1, "räikkönen": 1, 
		"sergio pérez": 1, "sergio perez": 1, "massa": 1, "adrian sutil": 1,
		"esteban gutierrez": 1, "esteban gutiérrez": 1, "grosjean": 1, 
		"kvyat": 1, "kobayashi": 1, "maldonado": 1, "chilton": 1,
		"bianchi": 1, "charlie whiting": 1, "villeneuve": 1, "schumacher": 1, 
		"newey": 1, "red bull": 1, "toro rosso": 1, "mc-laren": 1, "mercedes": 1,
		"mclaren": 1, "force india": 1, "ferrari": 1, "williams": 1, "sauber": 1, 
		"marusia": 1, "monoplaza": 1, "escuderia": 1, "escudería": 1, "caterham": 1, 
		"lotus": 1, 	
		// MotoGP
		"marc márquez": 1, "rossi": 1, "jorge lorenzo": 1, "pedrosa": 1,
		"fennati": 1, "rabat": 1, "viñales": 1, "miller": 1, "rins": 1, 
		"kallio": 1, "aegerter": 1, "corsi": 1, "salom": 1, "luthi": 1, 
		"terol": 1, "pons": 1, "alberto puig": 1, "angel nieto": 1, 
		"circuito": 1, "vuela rápida": 1, "dovizioso": 1, "espargaró": 1, 
		"espargaro": 1, "bradl": 1, "bautista": 1, "lannone": 1, "hayden": 1, 
		"aoyama": 1, "redding": 1, "crutchlow": 1, "barberá": 1, "barbera": 1, 
		// Other
		"sainz": 1, "marc coma": 1, "schlesser": 1, "ogier": 1, "kubica": 1,
		"peterhansel": 1, 
}

var Topics_sports =  map[string]int{
		"phelps": 1, "natación": 1, "natacion": 1, "atletismo": 1, "rugby": 1, 
		"esquí": 1, "padel": 1, "balonmano": 1, "golf": 1,	
		// Cycling words:
		"ciclismo": 1, "bici": 1, "bicicleta": 1, 
		"froome": 1, "contador": 1, "cadel evans": 1, "basso": 1, 
		"valverde": 1, "klöden": 1, "zubeldia": 1, "tour de francia": 1,
		"giro italia": 1, "vuelta españa": 1,
		"ag2r": 1, "radioshack": 1, "astana": 1, "lampre": 1,
}

var Topics_esports =  map[string]int{
		"starcraft": 1, "league of legends": 1, "lcs": 1, "dota": 1,
		"warcraft": 1, "hearthstone": 1, 
		"counter-strike": 1, "minecraft": 1, "gta": 1,
		"watchdogs": 1, "super mario": 1, "of isaac": 1, 
		"smite": 1, "battlefield": 1, "pokemon": 1, "final fantasy": 1, 
		"cod": 1, "of duty": 1, "the gathering": 1, "osu": 1, "of exile": 1,
		"fifa": 1, "dark souls": 1, "skyrim": 1, 	
		// LoL Teams
		"sk": 1, "roccat": 1, "gambit": 1, "nip": 1, "fnatic": 1, "supa hot": 1, "shc": 1, 
		"copenhagen wolves": 1, "alliance": 1, "millenium": 1, 
		"tsm": 1, "c9": 1, "evil geniuses": 1, "lmq": 1, "dignitas": 1, "curse": 1, "clg": 1,
		"complexity": 1, "denial": 1, "gamers2": 1, "h2k": 1, "unicorns of love": 1, "uol": 1,
		"skt": 1, "najin": 1, "cj entus": 1, "incredible miracle": 1, "jin air": 1, 
		"kt rolster": 1, "galaxy ozone": 1, "xenics": 1, "invictus gaming": 1, "ig": 1,
		"world elite": 1, "team we": 1, "omg": 1, "royal": 1, "lgd": 1, 
		"saigon jokers": 1, "azubu": 1, "taipei assassins": 1, "bangkok titans": 1, 
		// LoL Players
		"ocelote": 1, "xpeke": 1, "edward": 1, "morden": 1, "dioud": 1, "alex-ich": 1, 
		"alex ich": 1, "cyanide": 1, "yellowstar": 1, "soaz": 1, "rekkles": 1,
		"sjokz": 1, "nyph": 1, "tabzz": 1, "froggen": 1, "shook": 1, "wickd": 1, 
		"cowtard": 1, "youngbuck": 1, "darien": 1, "diamond": 1, "genja": 1,
		"creaton": 1, "jree": 1, "kerp": 1, "kottenx": 1, "kev1n": 1, 
		"celaver": 1, "xaxus": 1, "jankos": 1, "overpow": 1, "vander": 1, 
		"candy panda": 1, "svenskeren": 1, "jesiz": 1, "nrated": 1, "freddy122": 1, 
		"wewillfailer": 1, "impaler": 1, 
		"lemonnation": 1, "hai": 1, "balls": 1, "sneaky": 1, 
		"westrice": 1, "brokenshard": 1, "pr0lly": 1, "robertxlee": 1, "bubbadub": 1,
		"doublelift": 1, "aphromo": 1, "dexter": 1, "seraph": 1, 
		"voyboy": 1, "iwilldominate": 1, "quas": 1, "cop": 1, "xpecial": 1, 
		"krepo": 1, "altec": 1, "snoopeh": 1, "innox": 1, "pobelter": 1, 
		"xiaoweixiao": 1, "vasili": 1, "ackerman": 1, "noname": 1, "mor": 1, 
		"crumbzz": 1, "zion spartan": 1, "zionspartan": 1, "shiphtur": 1, "imaqtpie": 1, "kiwikid": 1,
		"wildturtle": 1, "bjergsen": 1, "dyrus": 1, "gleebglarbu": 1, 
		"flame": 1, "ambition": 1, "faker": 1, "madlife": 1, "toyz": 1, "lustboy": 1,
		"shy": 1, "ryu": 1, "insec": 1, "kakao": 1, "pray": 1, "imp": 1, "mata": 1, 
		"dade": 1, "bengi": 1, "piglet": 1, "poohmanduh": 1, 
		"pdd": 1, 
		// Spanish Starcraft players
		"lucifron": 1, "vortix": 1, "lolvsxd": 1, 
}

var Topics_music =  map[string]int{
		"rock": 1, "pop": 1, "indie": 1, "tango": 1,
		"folk": 1, "blues": 1, "jazz": 1, "rap": 1, "punk": 1,
		"balada": 1, "baladas": 1, "bolero": 1, "musica": 1, "música": 1,
		"country": 1, "dance": 1, "cumbia": 1, "tecno": 1, "electro": 1,
		"flamenco": 1, "gospel": 1, "hip hop": 1, "lambada": 1, "merengue": 1,
		"heavy": 1, "heavy-metal": 1, "reggaeton": 1, "samba": 1, "salsa": 1,
		// Artists
		"avicii": 1, 
}

var Topics_technology = map[string]int{
		"pc": 1, "informática": 1, "informatica": 1, "portatil": 1, "portátil": 1,
		"tablet": 1, "mac": 1, "iphone": 1, "android": 1, "samsung": 1, "galaxy": 1, 
		"tecnología": 1, "tecnologia": 1, "hardware": 1, "software": 1, "app": 1,
		"windows": 1, "linux": 1, "redhat": 1, "fedora": 1, "ubuntu": 1, "centos": 1,
		"debian": 1, "server": 1, "servidor": 1, "backtrack": 1, "gentoo": 1, 
		"suse": 1, "wifiway": 1, "usb": 1, "vga": 1, "hdmi": 1, "nokia": 1,
		"lg": 1, "xiaomi": 1, "lumia": 1, "nexus": 1, "google": 1, "facebook": 1,
		"api": 1, "db": 1, "bd": 1, "interfaz": 1, "interface": 1, 
		"wifi": 1, "router": 1, "wireless": 1, "bluetooth": 1, "impresora": 1,
		"cargador": 1, "teclado": 1, "mouse": 1, "keyboard": 1,
		"ios": 1, "python": 1, "golang": 1, "java": 1, "c++": 1, "oracle": 1, 
		"mongodb": 1, "shell": 1, "lisp": 1, "ada": 1, "prolog": 1, "ensamblador": 1, 
		"compilar": 1, "compilador": 1, "json": 1, "mysql": 1, "jquery": 1, 
		"javascript": 1, "php": 1,
}

var Topics_animals = map[string]int{
		"perro": 1, "gato": 1, "pájaro": 1, "pajaro": 1, "loro": 1,
		"pez": 1, "acuario": 1, "jaula": 1, "caballo": 1, "conejo": 1, "oveja": 1, 
		"iguana": 1, "araña": 1, "reptil": 1, "rana": 1, "tortuga": 1, 
		"periquito": 1, "ave": 1, "mamífero": 1, "mamifero": 1, "delfin": 1, "delfín": 1,
		"chimpancé": 1, "chimpance": 1, "león": 1, "tigre": 1, "jabalí": 1, "jabali": 1, "ballena": 1,
		"tiburón": 1, "vertebrado": 1, "invertebrado": 1, "ovíparo": 1, "oviparo": 1,
		"vivíparo": 1, "viviparo": 1, "oca": 1, "pato": 1, "ganso": 1, "cigüeña": 1, 
		"paloma": 1, "gaviota": 1, "pingüino": 1, 
		"patas": 1, "ocico": 1, "pezuña": 1, "ala": 1, "alas": 1, "pico": 1, 
		"zoología": 1, "animal": 1, "animales": 1, "mascota": 1,
}

var Topics_arts = map[string]int{
		"cuadro": 1, "exposición": 1, "museo": 1, "pintura": 1, 
		"pintor": 1, "artista": 1, "artistas": 1, "óleo": 1, "oleo": 1, 
		"comics": 1, "comic": 1, "dibujar": 1, "pintar": 1, "dibujo": 1, 
		"escultura": 1, "esculpir": 1, "escultor": 1, "baile": 1, 
		"bailarín": 1, "danza": 1, "pincel": 1, "lienzo": 1,
		"paleta": 1, "tallar": 1, "busto": 1, "acuarela": 1,
		"impresionista": 1, "expresionista": 1, "cubista": 1, 
		"naturalista": 1, "prerrafaelista": 1, "simbolista": 1, "modernista": 1,
		"surrealista": 1,
		"impresionismo": 1, "expresionismo": 1, "realismo": 1, "cubismo": 1, 
		"naïf": 1, "naif": 1, "naturalismo": 1, "gótico": 1, "románico": 1, 
		"renacentista": 1, "renacentismo": 1, "barroco": 1, "clasicismo": 1,
		"rococó": 1, "prerrafaelismo": 1, "simbolismo": 1, "modernismo": 1,
		"surrealismo": 1, 
		// People
		"picasso": 1, "giotto": 1, "leonardo": 1, "da vinci":1, "cézanne": 1,
		"cezanne": 1, "rembrandt": 1, "velázquez": 1, "kandinsky": 1, 
		"monet": 1, "caravaggio": 1, "van eyck": 1, 
		"durero": 1, "pollock": 1, "miguel ángel": 1, "gaugin": 1, "goya": 1,
		"van gogh": 1, "manet": 1, "matisse": 1, "rafael": 1, "basquiat": 1,
		"munch": 1, "tiziano": 1, "rubens": 1, "warhol": 1, "joan miró": 1, 
		"delacroix": 1, "el greco": 1, "degas": 1, "dalí": 1, 
		"tintoretto": 1, "botticelli": 1, "renoir": 1, "braque": 1, 
		"zurbarán": 1, "rossetti": 1, 
}

var Topics_cinema = map[string]int{
		"pelicula": 1, "cine": 1, "actor": 1, "actriz": 1, "película": 1, 
		"taquilla": 1, "productor": 1, "películas": 1, "peliculas": 1,
		"cortometraje": 1, "largometraje": 1,
		// People
		// Actors/Actresses:
		"liv tyler": 1, "melanie griffith": 1, "linda hamilton": 1, "denzel washington": 1,
		"john malkovich": 1, "matt damon": 1, "hugh grant": 1, "jim carrey": 1, "sean penn": 1, 
		"jeremy irons": 1, "jack nicholson": 1, "paz vega": 1, "val kilmer": 1, 
		"kim bassinger": 1, "milla jovovich": 1, "will smith": 1, "janet jackson": 1, 
		"lisa kudrow": 1, "christian slater": 1, "pamela anderson": 1, "drew barrymore": 1, 
		"lucy liu": 1, "morgan freeman": 1, "schwarzenegger": 1, "bruce willis": 1, 
		"ewan mcgregor": 1, "ethan hawke": 1, "kevin costner": 1, "victoria abril": 1, 
		"monica bellucci": 1, "sandra bullock": 1, "laetitia casta": 1, "penélope cruz": 1, 
		"salma hayek": 1, "jodie foster": 1, "love hewitt": 1, "gwyneth Paltrow": 1, 
		"bridget fonda": 1, "cámeron díaz": 1, "carmen electra": 1, "katherine zeta-jones": 1, 
		"julia roberts": 1, "jane fonda": 1, "jennifer aniston": 1, "jessica alba": 1, 
		"halle berry": 1, "demi moore": 1, "natalie portman": 1, "meg ryan": 1, "winona ryder": 1, 
		"uma thurman": 1, "maribel verdú":1, "sigourney weaver": 1, "tom cruise": 1, 
		"sean connery": 1, "antonio banderas": 1, "ben affleck": 1, "alec baldwin": 1, 
		"nicolas cage": 1, "de niro": 1, "benicio del": 1, "dicaprio": 1, 
		"douglas": 1, "clint eastwood": 1, "harrison ford": 1, "andy garcía": 1, "richard gere": 1,
		"mel gibson": 1, "anthony hopkins": 1, "paul newman": 1, "eduardo noriega": 1,
		"al pacino": 1, "brad pitt": 1, "john travolta": 1, "almodóvar": 1, 
		"robert redford": 1, "silvester stallone": 1, "bardem": 1, 
		"katharine hepburn": 1, "sofía loren": 1, "rob lowe": 1, "sharon stone": 1, 
		"david duchovny": 1, "brigitte bardot": 1, 
		// Directors
		"quentin tarantino": 1, "steven spielberg": 1, "kubrik": 1, "james cameron": 1, 
		"scorsese": 1, "tim burton": 1, "coppola": 1, "hitchcock": 1, "chaplin": 1, 
		"charlot": 1, "woody allen": 1, "ridley scott": 1, "polanski": 1, 
		"george lucas": 1, "buñuel": 1, "orson welles": 1, "brian de palma": 1, 
		"kurosawa": 1, "david lynch": 1, "billy wilder": 1, "oliver stone": 1, 
		"fellini": 1, "john ford": 1, "david lean": 1, "william wyler": 1, 
		"buster keaton": 1, "zhang yimou": 1, 
		// Movies
		"malditos bastardos": 1, "sin city": 1, "kill bill": 1, "pulp fiction": 1, 
		"reservoir dogs": 1, "munich": 1, "atrapame si": 1, "soldado ryan": 1,
		"jurassic park": 1, "parque jurásico": 1, "de schindler": 1, "indiana jones": 1,
		"ET": 1, "chaqueta metálica": 1, "el resplandor": 1, "naranja mecánica": 1,
		"espartaco": 1, "avatar": 1, "titanic": 1, "terminator": 1, "infiltrados": 1, "el aviador": 1,
		"gangs of": 1, "taxi driver": 1, "big fish": 1, "eduardo manostijeras": 1, 
		"origen": 1, "batman": 1, "memento": 1, "drácula": 1, "el padrino": 1, "apocalypse now": 1,
		"los pájaros": 1, "psicosis": 1, "ventana indiscreta": 1, 
		"dollar baby": 1, "mystic river": 1, "de madison": 1, 
		"candilejas": 1, "tiempos modernos": 1, "gran dictador": 1,
		"esdla": 1, "seven": 1, "benjamin button": 1,
		"match point": 1, "scoop": 1, "gladiator": 1, "alien": 1, "el pianista": 1, 
		"star wars": 1, "las galaxias": 1, "requiem por": 1, 
		"the wrestler": 1, "perro andaluz": 1, "viridiana": 1,
		"babel": 1, "ciudadano kane": 1, "misión imposible": 1, "mision imposible": 1, 
		"los intocables": 1, "siete samurais": 1, "crepúsculo": 1, "crepusculo": 1,
		"wall street": 1, "milla verde": 1, "cadena perpetua": 1,
		"doctor zhivago": 1, "lawrence de": 1, "río kwai": 1, "la diligencia": 1,
		"oliver twist": 1, "amelie": 1, "amélie": 1, "ben-hur": 1, 
		"moby dick": 1, "halcón maltés": 1, "llamado deseo": 1, "dagas voladoras": 1, 
		"sherlock holmes": 1, "luna nueva": 1, "casablanca": 1, "nosferatu": 1, "malcolm x": 1,
		"el golpe": 1, "gran torino": 1, "cisne negro": 1, "django desencadenado": 1,
		"slumdog millionaire": 1, "del fauno": 1, "toy story": 1, "shutter island": 1, 
		"miss sunshine": 1, "brokeback mountain": 1, "agora": 1, "caballero oscuro": 1,
		"los miserables": 1, "los vengadores": 1, "los simios": 1, "el ilusionista": 1,
		"de vendetta": 1, "sweeney todd": 1, "el hobbit": 1, "invictus": 1, 
		"apellidos vascos": 1, "cristina barcelona": 1, "millenium": 1, "mar adentro": 1, 
		"harry potter": 1, "el perfume": 1, "iron man": 1, 
		
}

// Literatura?!?

// AGE WordList DB
var WordsDB_age13to18 = map[string]int{
		"deberes": 1, "colegio": 1, "cole": 1, "instituto": 1,
		"insti": 1, "matemáticas": 1, "matematicas": 1, 
		"trabajo": 1, "graduación": 1, "graduacion": 1,
		"aburrido": 1, "aburrida": 1, "aburro": 1,
		"xD": 1, "novio": 1, "novia": 1, 
		"historia": 1, "inglés": 1, "ingles": 1,
		"lengua": 1, "química": 1, "quimica": 1,
		"biología": 1, "biologia": 1, "física": 1,
		"fisica": 1, "francés": 1, "frances": 1,
		"odio": 1, ";)": 1, ":p": 1, ";d": 1, "xp": 1,
		"<3": 1, ":d": 1, ":(": 1, "(:": 1, ":'(": 1,
		"lol": 1, "guapo": 1, "guapa": 1, "mono": 1,
		"clase": 1, "clases": 1, "cancion": 1, "canción": 1,
		"examen": 1, "trimestre": 1, "mejor amigo": 1, 
		"mejor amiga": 1,
		
}

var WordsDB_age19to22 = map[string]int{
		"estudiar": 1, "estudiando": 1,
		"campus": 1, "uni": 1, "universidad": 1,
		"biblioteca": 1, "biblio": 1, "examen": 1,
		"profesor": 1, "joder": 1, "jodido": 1, "jodida": 1,
		"borrachera": 1, "resaca": 1, "finales": 1,
		"vacaciones": 1, "prácticas": 1, "practicas": 1,
		"compañero": 1, "botellón": 1, "botellon": 1,
		"clase": 1, "clases": 1, "horario": 1, 
		"residencia": 1, "cuatrimestre": 1, "piso": 1,
		"charla": 1, "entrevista": 1, "conferencia": 1,
		"mierda": 1, "proyecto": 1, "proyectos": 1,
		"beber": 1, "salir": 1, "trabajo": 1,
		
}

var WordsDB_age23to29 = map[string]int{
		"trabajo": 1, "voluntario": 1, "horario": 1,
		"disfrutar": 1, "disfrutando": 1, "tomar algo": 1,
		"piso": 1, "casa": 1, "boda": 1, "días libres": 1,
		"dias libres": 1, "cerveza": 1, "cervezas": 1, 
		"vacaciones": 1, "celebrar": 1, "celebración": 1,
		"celebracion": 1, "cama": 1, "cenar": 1, "cena": 1,
		"oficina": 1, "recados": 1, "recado": 1, "marido": 1,
		"mujer": 1, "factura": 1, "facturas": 1, "pagar": 1,
		"pagado": 1, "pagada": 1, "pagadas": 1, "pagados": 1,
		"dinero": 1, "pagando": 1, "impuestos": 1, "alquiler": 1, 
		"beber": 1, "salir": 1, "vino": 1, "relajar": 1,
		"relax": 1, "compañia": 1, "casado": 1, "casada": 1, 
		"puesto": 1, "entrevista": 1, "curriculum": 1, "cv": 1,
		"experiencia": 1, "negocio": 1, "contratar": 1,
		"contratado": 1, "interesado": 1, "interesada": 1, 
		"empresa": 1, "preparado": 1, "ducha": 1, "siesta": 1,
		"mudarse": 1, "mudanza": 1, "mudo": 1, "deseos": 1,		
}

var WordsDB_age30to65 = map[string]int{
		"familia": 1, "hija": 1, "hijo": 1, "hijas": 1, 
		"hijos": 1, "reposición": 1, "reposicion": 1,
		"marido": 1, "bendito": 1, "bendita": 1, "bendiga": 1,
		"caracteristicas": 1, "características": 1,
		"dios": 1, "oracion": 1, "oraciones": 1, "rezar": 1,
		"rezo": 1, "rezaré": 1, "rezare": 1, "crios": 1, "niños":1,
		"entiendo": 1, "comprendo": 1, "adulto": 1, "padres": 1,
		"padre": 1, "madre": 1, "enseñar": 1, "enseño": 1,
		"joven": 1, "juventud": 1, "tiempo": 1, "estupendo": 1,
		"fantástico": 1, "comida": 1, "reunión": 1, "reunion": 1,
		"visitar": 1, "visitado": 1, "visitando": 1, "visito": 1,
		"país": 1, "pais": 1, "libertad": 1, "mujeres": 1, 
		"hombres": 1, "veteranos": 1, "recuerdo": 1, "recordar": 1,
		"orgulloso": 1, "orgullosa": 1, "orgullosos": 1, "orgullosas": 1,
		"mayor": 1, "mayores": 1, "vida": 1,	
}

// var WordsDB_age19to22 []WordWeight = []WordWeight{
		// WordWeight{Word: "estudiar", Weight: 1}, WordWeight{Word: "estudiando", Weight: 1},
		// WordWeight{Word: "campus", Weight: 1}, WordWeight{Word: "uni", Weight: 1},
		// WordWeight{Word: "universidad", Weight: 1}, WordWeight{Word: "semestre", Weight: 1},
		// WordWeight{Word: "biblioteca", Weight: 1}, WordWeight{Word: "biblio", Weight: 1},
		// WordWeight{Word: "examen", Weight: 1}, WordWeight{Word: "profesor", Weight: 1},
		// WordWeight{Word: "joder", Weight: 1}, WordWeight{Word: "jodido", Weight: 1},
		// WordWeight{Word: "jodida", Weight: 1},
		// WordWeight{Word: "borrachera", Weight: 1}, WordWeight{Word: "resaca", Weight: 1},
		// WordWeight{Word: "finales", Weight: 1}, WordWeight{Word: "vacaciones", Weight: 1},
		// WordWeight{Word: "prácticas", Weight: 1}, WordWeight{Word: "practicas", Weight: 1},
		// WordWeight{Word: "compañero", Weight: 1}, WordWeight{Word: "botellón", Weight: 1},
		// WordWeight{Word: "botellon", Weight: 1}, WordWeight{Word: "clase", Weight: 1},
		// WordWeight{Word: "clases", Weight: 1}, WordWeight{Word: "horario", Weight: 1},
		// WordWeight{Word: "residencia", Weight: 1}, WordWeight{Word: "cuatrimestre", Weight: 1},
		// WordWeight{Word: "piso", Weight: 1}, WordWeight{Word: "charla", Weight: 1},
		// WordWeight{Word: "entrevista", Weight: 1}, WordWeight{Word: "conferencia", Weight: 1},
		// WordWeight{Word: "mierda", Weight: 1}, WordWeight{Word: "proyecto", Weight: 1},
		// WordWeight{Word: "proyectos", Weight: 1}, WordWeight{Word: "beber", Weight: 1},
// }

// var WordsDB_age13to18 []WordWeight = []WordWeight{
		// WordWeight{Word: "deberes", Weight: 1}, 
		// WordWeight{Word: "colegio", Weight: 1}, WordWeight{Word: "cole", Weight: 1},
		// WordWeight{Word: "instituto", Weight: 1}, WordWeight{Word: "insti", Weight: 1},
		// WordWeight{Word: "matemáticas", Weight: 1}, WordWeight{Word: "matematicas", Weight: 1},
		// WordWeight{Word: "trabajo", Weight: 1},
		// WordWeight{Word: "graduación", Weight: 1}, WordWeight{Word: "graduacion", Weight: 1},
		// WordWeight{Word: "aburrido", Weight: 1},
		// WordWeight{Word: "aburro", Weight: 1},
		// WordWeight{Word: "xD", Weight: 1},
		// WordWeight{Word: "xd", Weight: 1},
		// WordWeight{Word: "novio", Weight: 1},
		// WordWeight{Word: "novia", Weight: 1},
		// WordWeight{Word: "historia", Weight: 1},
		// WordWeight{Word: "inglés", Weight: 1}, WordWeight{Word: "ingles", Weight: 1},
		// WordWeight{Word: "lengua", Weight: 1},
		// WordWeight{Word: "química", Weight: 1}, WordWeight{Word: "quimica", Weight: 1},
		// WordWeight{Word: "biología", Weight: 1}, WordWeight{Word: "biologia", Weight: 1},
		// WordWeight{Word: "física", Weight: 1}, WordWeight{Word: "fisica", Weight: 1},
		// WordWeight{Word: "francés", Weight: 1}, WordWeight{Word: "frances", Weight: 1},
		// WordWeight{Word: "odio", Weight: 1},
		// WordWeight{Word: ";)", Weight: 1}, WordWeight{Word: ":p", Weight: 1},
		// WordWeight{Word: ";d", Weight: 1}, WordWeight{Word: "xp", Weight: 1},
		// WordWeight{Word: "<3", Weight: 1}, WordWeight{Word: ":d", Weight: 1},
		// WordWeight{Word: ":(", Weight: 1}, WordWeight{Word: "(:", Weight: 1},
		// WordWeight{Word: ":'(", Weight: 1}, WordWeight{Word: "lol", Weight: 1},
		// WordWeight{Word: "guapo", Weight: 1}, WordWeight{Word: "guapa", Weight: 1},
		// WordWeight{Word: "mono", Weight: 1}, WordWeight{Word: "horario", Weight: 1},
		// WordWeight{Word: "clase", Weight: 1}, WordWeight{Word: "clases", Weight: 1},
		// WordWeight{Word: "canción", Weight: 1}, WordWeight{Word: "cancion", Weight: 1},
		// WordWeight{Word: "examen", Weight: 1}, WordWeight{Word: "trimestre", Weight: 1},
		// Two words
		// WordWeight{Word: "mejor amigo", Weight: 1}, WordWeight{Word: "mejor amiga", Weight: 1},
// }