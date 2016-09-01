package trendings

import (
	"sort"
	"strings"
	//"github.com/araddon/httpstream"
	"labix.org/v2/mgo"			// MongoDB Driver 
	"labix.org/v2/mgo/bson" 	// MongoDB BSON translator
	"time"
	"twittertypes"
	"fmt"
)

var BlackList = []string {"a", "ante", "bajo", "cabe", "con", "contra", "de", "desde", "en", "entre", "hacia",
								"hasta", "para", "por", "segun", "sin", "so", "sobre", "tras", "mediante", "durante", // Preposiciones
								"el", "la", "los", "las", "un", "una", "unos", "unas", "este", "ese", "aquel",
								"esta", "esa", "aquella", "estos", "esos", "aquellos", "estas", "esas", "aquellas", // Determinantes
								"y", "e", "o", "u", "mas", "ni", "sino", "siquiera", "aunque", "como", "conque",
								"cuando", "donde", "entonces", "ergo", "incluso", "luego", "mientras", "porque", 
								"pues", "que", "sea", "ya", // Conjunciones
								"yo", "tu", "el", "ella", "nosotros", "nosotras", "vosotros", "vosotras", "ellos", "ellas", // Pronombres
								"del", "me", "te", "se", "nos", "os", "lo", "le", "mi", "ti", "si",
								"que", "quien", "cual", "cuanto", "cuando", "donde", "como",
								"mio", "tuyo", "suyo", "nuestro", "vuestro", "mia", "tuya", "suya", "nuestra", "vuestra",
								"pero", // Adverbios
								"ha", "he", "hemos", "habeis", "han", "está", "iban", // Verbos
								"rt", "http", "t", "co", "pic", "twitter", "com", "es", "net", "org", "vine", "goo", // Twitter 
								"no", "hoy", "al", "-",// Especiales
								// De la observacion de Twitter
								"más", "qué", "aquí", "todo", "su",
								}

// IsBlackListed returns TRUE if the word 'word' is in the list 'blacklist', returns FALSE in any other case
func IsBlackListed (blacklist []string, word string) bool {
	for i:=0; i<len(blacklist); i++ {
		if blacklist[i] == word {
			return true
		}
	}
	return false
}

// AddWord adds 'word' to the map 'm', if 'word' already exists, it increments its counter
func AddWord(m map[string]int, word string) {
	if word == "" {
		return
	}
	if len(word) < 2 {
		return 
	}
	if IsBlackListed(BlackList, word) == true {
		return
	}
	value, exist := m[word]
	if exist == false {
		m[word] = 1
	} else {
		m[word] = value+1
	}
}

type Pair struct {
	Word string
	Value int
}

// A slice of Pairs that implements sort.Interface to sort by Value.
type WordsList []Pair
func (p WordsList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p WordsList) Len() int { return len(p) }
func (p WordsList) Less(i, j int) bool { return p[i].Value > p[j].Value } // in this case >

// A function to turn a map into a PairList, then sort and return it. 
func sortMapByValue(m map[string]int) WordsList {
   p := make(WordsList, len(m))
   i := 0
   for k, v := range m {
      p[i] = Pair{k, v}
	  i++
   }
   sort.Sort(p)
   return p
}

type Fav struct {
	User		VerifiedUser
	Value		int64
}
// A slice of Pairs that implements sort.Interface to sort by Value.
type ListaFavs []Fav
func (p ListaFavs) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p ListaFavs) Len() int { return len(p) }
func (p ListaFavs) Less(i, j int) bool { return p[i].User.Value > p[j].User.Value } // in this case >

// A function to turn a map into a PairList, then sort and return it. 
func SortMapByValue2(m map[int64]VerifiedUser) ListaFavs {
   p := make(ListaFavs, len(m))
   i := 0
   //var demInt int64 = 0
   for k, v := range m {
      p[i] = Fav{v, k}
	  i++
   }
   sort.Sort(p)
   return p
}

// TweetToText gets a tweet and returns its text but with some URLs filtered
func TweetToText (tweet twittertypes.Tweet) string { 
	var index []int
	index = make([]int, 2)
	index[0] = 139
	index[1] = 140
	text := tweet.Text
	for i:=0; i<len(tweet.Entities.Urls); i++ {
		text = strings.Replace(text, tweet.Entities.Urls[i].Url, " ", -1)
	}
	for i:=0; i<len(tweet.Entities.Media); i++ {
		text = strings.Replace(text, tweet.Entities.Media[i].Url, " ", -1)
	}
	return text
}

func WordsToDoubleWords (words []string) []string {
	var doubleWords []string
	//doubleWords = make([]string, len(words))
	for i:=1; i<len(words); i++ {
		//doubleWords[i-1] = strings.Join(words[i-1:i+1], " ")
		doubleWords = append(doubleWords, strings.Join(words[i-1:i+1], " "))
	}
	return doubleWords
}

func EmptyWordsFilter (words []string) []string {
	var filteredWords []string
	for i:=1; i<len(words); i++ {
		if words[i] != "" {
			filteredWords = append(filteredWords, words[i])
		}
	}
	return filteredWords
}

func WordsFilter (words []string, hashtag string) []string {
	var filteredWords []string
	for i:=1; i<len(words); i++ {
		if ((words[i] != "") && (words[i] != hashtag)) && (IsBlackListed(BlackList, words[i]) == false) {
			filteredWords = append(filteredWords, words[i])
		}
	}
	return filteredWords
}

func DoubleWordsFilter (words []string, hashtag string) []string {
	var filteredWords []string
	var blacklisted bool = false
		for i:=1; i<len(words); i++ {
			if (strings.Contains(words[i], hashtag) == false) || (hashtag == "") {
				singles := strings.Split(words[i], " ")
				for j:=0; j<len(singles); j++ {
					if IsBlackListed(BlackList, singles[j]) == true {
						blacklisted = true
					}
				}
				if blacklisted == false {
					filteredWords = append(filteredWords, words[i])
				}
				blacklisted = false
			} 
		}
	return filteredWords
}

// TextToWords splits the text into words with some "filters"
func TextToWords (t string) []string {
 	t = strings.Replace(t, ".", " ", -1)
	t = strings.Replace(t, ",", " ", -1)
	t = strings.Replace(t, ":", " ", -1)
	t = strings.Replace(t, "!", " ", -1)
	t = strings.Replace(t, "?", " ", -1)
	t = strings.Replace(t, "\n", " ", -1)
	t = strings.Replace(t, ")", " ", -1)
	t = strings.Replace(t, "(", " ", -1)
	t = strings.Replace(t, "[", " ", -1)
	t = strings.Replace(t, "]", " ", -1)
	t = strings.Replace(t, "/", " ", -1) 
	t = strings.Replace(t, "…", " ", -1)
	t = strings.Replace(t, "¡", " ", -1)
	t = strings.Replace(t, "¿", " ", -1)
	t = strings.Replace(t, "\"", " ", -1) // "\""
	t = strings.Replace(t, "“", " ", -1)
	t = strings.Replace(t, "”", " ", -1)
	t = strings.ToLower(t)
	words := strings.Split(t, " ")
	return words
} 

// 
func MergeTrendings (words WordsList, doubleWords WordsList) WordsList{
	var mergedList WordsList
	mergedList = make(WordsList, 100)
	var singles WordsList
	singles = make(WordsList, 100)
	for i:=0; i<100; i++ {
		singles[i] = words[i]
	}
	var doubles WordsList
	doubles = make(WordsList, 100)
	for i:=0; i<100; i++ {
		doubles[i] = doubleWords[i]
	}
	
	// fmt.Println("Simples")
	// for i:=0; i<20; i++ {
		// fmt.Println(singles[i])
	// }
	
	// fmt.Println("Dobles:")
	// for i:=0; i<20; i++ {
		// fmt.Println(doubles[i])
	// }
	// fmt.Println("")
	
	for i:=0; i<20; i++ {
		for j:=0; j<7; j++ {
			if strings.Contains(doubles[j].Word, singles[i].Word) {
				singles[i].Value = singles[i].Value - doubles[j].Value
			}
		}
	}
	sort.Sort(singles)
	sort.Sort(doubles)

	indice1 := 0
	indice2 := 0
	for i:=0; i<20; i++ {
		if singles[indice1].Value > doubles[indice2].Value {
			mergedList[i] = singles[indice1]
			indice1++
		} else {
			mergedList[i] = doubles[indice2]
			indice2++
		}
	}
	// fmt.Println("Merged:")
	// for i:=0; i<20; i++ {
		// fmt.Println(mergedList[i])
	// }
	fmt.Println("")
	return mergedList
}

// ParseTweetsAll receives a list of tweets and returns a list of pairs [word, value]
func ParseTweetsAll (tweets []twittertypes.Tweet, hashtag string) (WordsList, WordsList) {
// Procesamos los tweets y devolvemos la lista procesada
	words := make(map[string]int)	
	doubleWords := make(map[string]int)
	var text string
	var result []string
	var doubleResult []string
	var sortedWords WordsList
	var sortedDoubleWords WordsList
	for i:= 0; i < len(tweets); i++ {
		text = TweetToText(tweets[i])		
		result = TextToWords(text)									// Text split into Words
		result = EmptyWordsFilter(result)							// Filtered empty Words
		doubleResult = WordsToDoubleWords(result)					// Words merged into DoubleWords
		result = WordsFilter(result, hashtag)						// Words filtered by hashtag
		doubleResult = DoubleWordsFilter(doubleResult, hashtag)		// DoubleWords filered by hashtag
		
		// We place the words into a map
		for j:=0; j< len(result); j++ {
			AddWord(words, result[j])
		}
		for j:=0; j< len(doubleResult); j++ {
			AddWord(doubleWords, doubleResult[j])
		}	
	
	}
	
	sortedWords = sortMapByValue(words)
	sortedDoubleWords = sortMapByValue(doubleWords)
	return sortedWords, sortedDoubleWords
}

// ParseTweetsKeyword receives a list of tweets, filters it by 'keyword' and returns a list of pairs [word, value]
func ParseTweetsKeyword (tweets []twittertypes.Tweet, keyword string, hashtag string) (WordsList, WordsList) {
	// If the tweet does not contain the 'keyword', we discard it
	// and we return the processed list
	words := make(map[string]int)	
	doubleWords := make(map[string]int)
	var text string
	var result []string
	var doubleResult []string
	var sortedWords WordsList
	var sortedDoubleWords WordsList
	for i:=0; i< len(tweets); i++ {
		text = TweetToText(tweets[i])		
		result = TextToWords(text)									// Text split into Words
		result = EmptyWordsFilter(result)							// Filtered empty Words
		doubleResult = WordsToDoubleWords(result)					// Words merged into DoubleWords
		result = WordsFilter(result, hashtag)						// Words filtered by hashtag
		doubleResult = DoubleWordsFilter(doubleResult, hashtag)		// DoubleWords filtered by hashtag
		doubleResult = DoubleWordsFilter(doubleResult, keyword)		// DoubleWords filtered by keyword
		if ContainsWord(result, keyword) == true {
			for j:=0; j< len(result); j++ {
				AddWord(words, result[j])
			}
			for j:=0; j< len(doubleResult); j++ {
				AddWord(doubleWords, doubleResult[j])
			}	
		}
	}
	delete(words, keyword)
	sortedWords = sortMapByValue(words)
	sortedDoubleWords = sortMapByValue(doubleWords)
	return sortedWords, sortedDoubleWords
}

// ContainsWord returns true if 'word' is one of the word's list, false otherwise
func ContainsWord (list []string, word string) bool {
	for i:=0; i<len(list); i++ {
		if list[i] == word {
			return true
		}
	}
	return false
}

// GetTweetsRange gets from the collection the tweets în the time period determined by [start, end]
func GetTweetsRange (colec mgo.Collection, start, end time.Time) []twittertypes.Tweet {
	var tweetsDate []twittertypes.Tweet
	err := colec.Find(bson.M{"epoch": bson.M{ "$gte": start.Unix(), "$lt": end.Unix()}}).All(&tweetsDate)
	if err != nil { 
		panic(err) 
	}	
	return tweetsDate
}


// GetTweetsKeyword obtains the tweets containing 'keyword' in the collection
func GetTweetsKeyword (colec mgo.Collection, keyword string) []twittertypes.Tweet {
	var tweetsKeyword []twittertypes.Tweet
	var tweetsAll []twittertypes.Tweet
	var result []string
	tweetsAll = GetTweetsAll(colec)
	for i:=0; i< len(tweetsAll); i++ {
		result = TextToWords(tweetsAll[i].Text)
		if ContainsWord(result, keyword) == true {
			tweetsKeyword = append(tweetsKeyword, tweetsAll[i])
		}
	}
	return tweetsKeyword
}

// GetTweetsVerified obtains all the Tweets with Verified User from a collection 
func GetTweetsVerified (colec mgo.Collection) []twittertypes.Tweet {
	var tweetsVerified []twittertypes.Tweet
	err := colec.Find(bson.M{"user.verified": true}).All(&tweetsVerified)
	if err != nil { 
		panic(err) 
	}	
	return tweetsVerified
}

// GetTweetsAll gets all the tweets from the collection
func GetTweetsAll (colec mgo.Collection) []twittertypes.Tweet {
	var tweets []twittertypes.Tweet
	// Lectura de los tweets de la BD
	err := colec.Find(bson.M{"text": bson.M{"$exists": true}}).All(&tweets)
	if err != nil { 
		panic(err) 
	}
	return tweets
}

// GetIntervalsN returns a list of 'n' almost-equal Intervals in the [start, end] Interval (the last interval can be bigger)
func GetIntervalsN (start, end time.Time, n int64) []Interval {
	var intervals []Interval 
	intervals = make([]Interval, n)
	var totalLength int64
	totalLength = end.Unix()-start.Unix()
	intervalLength := totalLength / n
	intervals[0].Start = start
	intervals[0].End = time.Unix(start.Unix() + intervalLength, 0)
	var i int64
	for i=1; i<n; i++ {
		intervals[i].Start = time.Unix(intervals[i-1].End.Unix(), 0)
		intervals[i].End = time.Unix(intervals[i].Start.Unix() + intervalLength , 0)
	}
	intervals[n-1].End = end 
	return intervals
}

// GetIntervalsLen returns a list of intervals of 'len' duration (the last interval duration could be smaller than the others)
func GetIntervalsLen (start, end time.Time, length time.Duration) []Interval {
	var intervals []Interval 
	totalLength := end.Unix()-start.Unix()
	intervalLength := int64(length.Seconds())
	n := (totalLength / intervalLength )+1
	intervals = make([]Interval, n)
	intervals[0].Start = start
	intervals[0].End = time.Unix(start.Unix() + intervalLength, 0)
	var i int
	for i=1; intervals[i-1].End.Before(end); i++ {
		intervals[i].Start = time.Unix(intervals[i-1].End.Unix(), 0)
		intervals[i].End = time.Unix(intervals[i].Start.Unix() + intervalLength , 0)
	}
	intervals[i-1].End = end 
	/* In case the intervals are exact match, we fix the length of the returned list */
	t := make([]Interval, i)
	copy(t, intervals)
	intervals = t
	/*                            */
	return intervals
}

type Interval struct {
	Start 	time.Time
	End 	time.Time
}

// ListVerifiedUsers returns a map [VerifiedUser, numTweets] of the Verified Users and their number of tweets in the list
func ListVerifiedUsers (tweets []twittertypes.Tweet) map[int64]VerifiedUser { //[]VerifiedUser {
	//var users []VerifiedUser
	verifiedsMap := make(map[int64]VerifiedUser)
	var tempUser VerifiedUser
	for i:=0; i<len(tweets); i++ {
		if tweets[i].User.Verified == true {
			// We update de user in the map (add or +1)
			if val, exists :=verifiedsMap[tweets[i].User.Id]; exists {
				tempUser.Id = tweets[i].User.Id
				tempUser.Screen_name = tweets[i].User.Screen_name
				tempUser.Value = val.Value+1
				verifiedsMap[tweets[i].User.Id] = tempUser
			} else {
				tempUser.Id = tweets[i].User.Id
				tempUser.Screen_name = tweets[i].User.Screen_name
				tempUser.Value = 1
				verifiedsMap[tweets[i].User.Id] = tempUser;
			}
		}
	}
	return verifiedsMap
	//return users
}

type VerifiedUser struct {
	Id 				int64
	Screen_name		string
	Value 			int
}


