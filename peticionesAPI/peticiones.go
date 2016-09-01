package main

import "fmt"
import "net/http"
import "io/ioutil"

func main () {
for i:=0; i< 10000; i++ {
	resp, err := http.Get("http://api.genderize.io?name=Manu")
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()	
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

}