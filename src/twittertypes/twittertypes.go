package twittertypes

import (
	"bytes"
	"time"
)

type UserProfile struct {
	Screen_name								string
	Favourites_count						int
	Profile_background_color				string
	Profile_background_image_url_https		string
	Profile_background_tile					bool
	Profile_image_url						string
	Profile_use_background_image			bool 
	Id_str									string 		
	Name									string
	Friends_count							int 
	Utc_offset								int
	Geo_enabled								bool
	Is_translator							bool
	Profile_link_color						string 
	Default_profile							bool 
	Default_profile_image					bool
	//Follow_request_sent						bool
	//Notifications							bool		 
	Location								string 
	Profile_sidebar_border_color			string 
	//Following								bool 
	Entities								Entities
	//Is_translation_enabled					bool 
	Profile_image_url_https					string
	Description								string 
	Protected								bool 
	Profile_background_image_url			string
	Profile_sidebar_fill_color				string 
	Url										string 
	Created_at								string
	Time_zone								string 
	Verified								bool 
	Lang									string
	Contributors_enabled					bool 
	Id										int64 
	Followers_count							int 
	Listed_count							int 
	Statuses_count							int 
	Status									Tweet 
	Profile_text_color						string
}

type User struct {
	Lang                         string
	Verified                     bool
	Followers_count              int
	Location                     string
	Screen_name                  string
	Following                    bool
	Friends_count                int
	Profile_background_color     string
	Favourites_count             int
	Description                  string
	Notifications                string
	Profile_text_color           string
	Url                          string
	Time_zone                    string
	Statuses_count               int
	Profile_link_color           string
	Geo_enabled                  bool
	Profile_background_image_url string
	Protected                    bool
	Contributors_enabled         bool
	Profile_sidebar_fill_color   string
	Name                         string
	Profile_background_tile      string
	Created_at                   string
	Profile_image_url            string
	Id                           int64
	Utc_offset                   int
	Profile_sidebar_border_color string
}

type UserEncuesta struct {
	Verified                     bool
	Followers_count              int
	Location                     string
	Screen_name                  string
	Friends_count                int
	Statuses_count               int
	Name                         string
	Created_at                   string
	Id                           int64
}

type Request struct {
	User       string
	Timestamp  time.Time
	Smartpoint string
	City       string
	Msg        string
	Option     string
	//location: {type:"Point", coordinates:[]},
	Status string
}

type Entities struct {
	Hashtags []struct {
		Indices []int
		Text    string
	}
	Urls []struct {
		Indices      []int
		Url          string
		Display_url  string
		Expanded_url string
	}
	User_mentions []struct {
		Name        string
		Indices     []int
		Screen_name string
		Id          int64
		Id_str      string
	}
	Media []struct {
		Id              int64
		Id_str          string
		Media_url       string
		Media_url_https string
		Url             string
		Display_url     string
		Expanded_url    string
		Sizes           MediaSizes
		Type            string
		Indices         []int
	}
}
type MediaSizes struct {
	Medium MediaSize
	Thumb  MediaSize
	Small  MediaSize
	Large  MediaSize
}

type MediaSize struct {
	W      int
	H      int
	Resize string
}

type TweetEncuesta struct {
	Text                    string
	Geo                     Coords
	Id                      int64
	User                    *UserEncuesta
	Epoch					int64
	Lang					string
	Created_at				string
}

type Tweet struct {
	Text                    string
	Truncated               bool
	Geo                     Coords
	In_reply_to_screen_name string
	Favorited               bool
	Source                  string
	Contributors            string
	In_reply_to_status_id   string
	In_reply_to_user_id     int64
	Id                      int64
	Created_at              string
	User                    *User
	Entities                *Entities
	Epoch					int64
	//
	Coordinates				Coords	
	Retweet_count			int
	Favorite_count			int
	Retweeted_status 		Tweet2
	Id_str					string
	Filter_level			string
	In_reply_to_status_id_str	string
	In_reply_to_user_id_str 	string
	Lang					string
	Possibly_sensitive		bool
	Retweeted				bool
	Withheld_copyright		bool
	Withheld_in_countries	[]string
	Withheld_scope			string
}

type Tweet2 struct {
	Text                    string
	Truncated               bool
	Geo                     Coords
	In_reply_to_screen_name string
	Favorited               bool
	Source                  string
	Contributors            string
	In_reply_to_status_id   string
	In_reply_to_user_id     int64
	Id                      int64
	Created_at              string
	User                    *User
	Entities                *Entities
	Epoch					int64
	//
	Coordinates				Coords	
	Retweet_count			int
	Favorite_count			int
	//Retweeted_status 		Tweet2
	Id_str					string
	Filter_level			string
	In_reply_to_status_id_str	string
	In_reply_to_user_id_str 	string
	Lang					string
	Possibly_sensitive		bool
	Retweeted				bool
	Withheld_copyright		bool
	Withheld_in_countries	[]string
	Withheld_scope			string
}

type Coords struct {
	Coordinates				[]float64
	Type					string
}

/*
The twitter stream contains non-tweets (deletes)

{"delete":{"status":{"user_id_str":"36484472","id_str":"191029491823423488","user_id":36484472,"id":191029491823423488}}}
{"delete":{"status":{"id_str":"191184618165256194","id":191184618165256194,"user_id":355665960,"user_id_str":"355665960"}}}
{"delete":{"status":{"id_str":"172129790210482176","id":172129790210482176,"user_id_str":"499324766","user_id":499324766}}}
{"delete":{"status":{"user_id_str":"366839894","user_id":366839894,"id_str":"116974717763719168","id":116974717763719168}}}
{"delete":{"status":{"user_id_str":"382739413","id":191184546841112579,"user_id":382739413,"id_str":"191184546841112579"}}}
{"delete":{"status":{"user_id_str":"388738304","id_str":"123723751366987776","id":123723751366987776,"user_id":388738304}}}
{"delete":{"status":{"user_id_str":"156157535","id_str":"190608148829179907","id":190608148829179907,"user_id":156157535}}}

*/
// a function to filter out the delete messages
func OnlyTweetsFilter(handler func([]byte)) func([]byte) {
	delTw := []byte(`{"delete"`)
	return func(line []byte) {
		if !bytes.HasPrefix(line, delTw) {
			handler(line)
		}
	}
}
