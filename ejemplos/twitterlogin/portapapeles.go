func TimelinesHandler(rw http.ResponseWriter, req *http.Request) {
	var (
	//	url       string
		err       error
		sessionID string
	)
	if _, err = GetSessionID(req); err != nil {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		http.Error(rw, "Log In before doing stuff", 400)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	if  _, prs := sessions[sessionID]; !prs {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return
	}
	sessionID, err = GetSessionID(req)
	if sessionID == "" {
		log.Printf("Got a query for stuff with no session id: %v\n", err)
		fmt.Fprintf(rw, "Debes loguearte con Twitter para poder consultar cosas</br>")
		fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
		return 
	}
	t, err := template.New("encabezado").Parse("Let's read some timelines!\n\n")
	err = t.ExecuteTemplate(rw, "encabezado", nil)
	fmt.Fprintf(rw, "<a href=\"/\">Back to main</a></br>")
	fmt.Fprintf(rw, "Sesion ID: " + sessionID)
	fmt.Println("SessionID =", sessionID)
	rw.(http.Flusher).Flush()
	
	// Connecting to mongo DB
	session, err := mgo.Dial("130.206.83.133") 
	if err != nil { 
		panic(err) 
	} 
	defer session.Close() 
	var user = "manupruebas"
	c := session.DB("timelines").C(user)
	profiles.GetTimelineToDB(*c, user, 500)
	todosLosTweets := trendings.GetTweetsAll(*c)
	for i:=0; i<len(todosLosTweets); i++ {
		//fmt.Println(todosLosTweets[i].Text)
		fmt.Fprintf(rw, todosLosTweets[i].Text + "\n")
	}
	rw.(http.Flusher).Flush()
	err = c.DropCollection()
	if err != nil { 
		panic(err) 
	}	
	fmt.Fprintf(rw, "\n\n\n")
	profiles.GetTimelineToDB(*c, user, 500)
	todosLosTweets2 := trendings.GetTweetsAll(*c)
	for i:=0; i<len(todosLosTweets2); i++ {
		//fmt.Println(todosLosTweets[i].Text)
		fmt.Fprintf(rw, todosLosTweets2[i].Text + "\n")
	}
	err = c.DropCollection()
	if err != nil { 
		panic(err) 
	}	
	return
}

