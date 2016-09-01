Fichero "perfiles.go" 

v1
- Funciones de: obtención de timeline y escritura en BD, cálculo de tweets cada día y media diaria
	* GetTimelineToDB --> Obtiene una timeline y la guarda en la DB (con maxTweets)
	* GetTweetsPerDay --> Lee una timeline de BD y saca una lista de tweets por día, media, etc.

v2
- Funciones de obtener perfiles de usuario y género
	* GetUserByScreenName --> Obtiene un perfil de usuario a partir de su 'screen_name'
	* GetUserByID --> Obtiene un perfil de usuario a partir de su 'id'
	* GenderFromName --> Calcula el género de una persona a partir de un nombre

v3
- Funciones de obtener Friends (following) y followers de un usuario
	* GetFriendsByIDToDB --> Obtiene los 'friends' de un usuario identificado por su Id y los guarda en BD
	* GetFriendsByScreenNameToDB --> Obtiene los 'friends' de un usuario identificado por su 'screenName' 
					 y los guarda en BD
	* GetFollowersByIDToDB --> Obtiene los 'followers' de un usuario identificado por su Id y los guarda en BD
	* GetFollowersByScreenNameToDB --> Obtiene los 'followers' de un usuario identificado por su 'screenName' 
					 y los guarda en BD

v4
- Creado el paquete "profiles" en %GOPATH%/src y movido el código de la nueva libreria

v5
- Funciones de: Obtener pais/provincia, Obtener localización del usuario
	* GetCountryAndStateByCoordinates --> Obtiene el pais y la provincia de las coordenadas de la geolocalizacion
						de un tweet
	* GetLocation --> Obtiene la localización de un usuario de twitter

v6
- Nueva versión de "profiles.go" modificada para la otra parte.

v7
- Añadido el control de limitaciones y errores de las distintas APIs a las siguientes funciones:
	*getuserbyscreenname, getuserbyid, getcountryandstatebycoordinates, genderfromname
- Añadida la URL a la que llamamos en el mensaje de "N calls available" en la API de twitter

v8
- Añadidos mapas de palabras para tratar de determinar la Edad (de momento no predice bien)

v9
- Añadidos mapas de intereses para determinar los intereses de una timeline.

v10
- Añadida la función para que ordene los maxRt y maxFav en el top5
