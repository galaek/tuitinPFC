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