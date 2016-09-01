package tweel_in

import (
	"bitbucket.org/548017/go_mobydick"
	"encoding/json"
	"github.com/548017/stmtlk"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Feeling string

func valorar(t go_mobydick.Tweet) stmtlk.Valoracion {
	api := stmtlk.NewStmtlkApi(apiculturAccessToken)
	s := strings.Replace(t.Text, "http://t.co/", "", -1)
	s = strings.Replace(s, "/", " ", -1)
	val, err := api.GetValoracion(s)
	if err != nil {
		panic(err)
	}
	return val
}

func bitextAnalyze(t go_mobydick.Tweet, client *http.Client) Feeling {
	lang := "ESP"
	service := "http://svc8.bitext.com/WS_Nops_Val/Service.aspx"
	data := url.Values{}
	data.Set("User", bitextUser)
	data.Set("Pass", bitextPass)
	data.Set("Lang", lang)
	data.Set("ID", "0")
	data.Set("Text", t.Text)
	data.Set("Detail", "Global")
	data.Set("OutFormat", "JSON")
	data.Set("Normalized", "Both")
	resp, err := client.PostForm(service, data)
	if err != nil {
		log.Print(err)
		return "fail"
	}

	if resp.StatusCode != 200 {
		p, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		log.Printf("Get %s returned status %d, %s", resp.Request.URL, resp.StatusCode, p)
		return "error"
	}
	p, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	var dat map[string]interface{}
	if err := json.Unmarshal(p, &dat); err != nil {
		panic(err)
	}
	datos := dat["data"].([]interface{})
	minidat := datos[0].(map[string]interface{})

	log.Println(minidat["global_value"].(float64))
	return "termine"
}

func analyze2(t go_mobydick.Tweet) Feeling {
	pos := []string{
		"great", "superb",
		"thank", "gracias",
		"good", "happy", "placer",
		"alegr", "bien",
		"nice", "buen",
		"amazing", "genial",
		"awesome", "increible",
		"wonderful", "maravilloso",
		"fantastic", "fantástico",
		"best", "mejor",
		"finally", "por fin",
		"perfect",
		"like", "gust",
		"brilliant", "brillante",
		"phenomenal", "fenomenal",
		"excellent", "excelente",
		"spectacular", "bad ass", "badass",
		"enjoy", "disfrut",
		"rapid", "rápid", "fast",
		"love", "encant", "luv",
		"content", "feliz",
		"congrat", "felicidades",
		"=)",
		":)", "xd", ":d", ":p", ":-)", ":o)", ":]", "8)", ":}", "8-D", "8D"}
	neg := []string{"bad", "malo",
		"horrible", "mierda",
		"hate", "odio", "odia",
		"terrible", "awful", "atroz",
		"suck", "apesta",
		"worst", "peor",
		"slow", "lent",
		"aburri", "bor",
		"dislike", "disgust",
		"hurt", "dolor", "duele", "dolid",
		"decepcion", "decepción", "disappoint",
		"detestable",
		"enem", "no mola", "no me gusta", "no es agradable", "no es divertido",
		":(", ":'(", ":s"}

	val := 0

	data := strings.ToLower(t.Text)
	data = strings.Replace(data, "http://t.co/", "", -1)

	for _, w := range pos {
		if strings.Contains(data, w) {
			val++
		}
	}
	for _, w := range neg {
		if strings.Contains(data, w) {
			val--
		}
	}
	if val > 0 {
		return "POSITIVA"
	} else if val < 0 {
		return "NEGATIVA"
	} else {
		return "NEUTRA"
	}
	return "NEUTRA"
}

func analyze(t go_mobydick.Tweet) Feeling {
	pos := []string{"great", "genial",
		"good", "bueno",
		"amazing", "genial",
		"awesome", "increible",
		"wonderful", "maravilloso",
		"fantastic", "fantastico",
		"best", "mejor",
		"perfect", "perfecto"}
	neg := []string{"bad", "malo",
		"horrible", "mierda",
		"hate", "odio",
		"terrible",
		"worst", "peor"}

	var f Feeling

	data := strings.Fields(t.Text)
	for _, w := range data {
		for _, wa := range pos {
			if strings.EqualFold(w, wa) {
				f = "POSITIVA"
			}
		}
		for _, wa := range neg {
			if strings.EqualFold(w, wa) {
				f = "NEGATIVA"
			}
		}
	}
	return f
}
