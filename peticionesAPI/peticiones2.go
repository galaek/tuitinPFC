package main

import "fmt"
import "net/http"
import "net/url"
import "io/ioutil"

func main () {
for i:=0; i< 1000; i++ {
	resp, err := http.PostForm("http://www.i-gender.com/ai", url.Values{"name": {"Manu"}})
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()	
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}

}