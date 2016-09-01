v1
- Lee tweets del stream de twitter y los guarda en la bd
- Saca los tweets de la BD y cuenta el numero de ocurrencias de cada palabra
- Guarda una lista ordenada de pares {palabra, ocurrencias} en la BD
- Saca la lista de pares y la muestra

v2
- Añadida la funcion de filtrar los tweets por una palabra
  Usada para sacar los subtrendings una vez obtenidos los trendings

v3
- Añadido un trozo de código de pruebas para los "time" de twitter y así
  saber que tweets estan en un rango de tiempo

v4
- Versión con las 4 funciones clave y un "mini-ejemplo" para probarlas:
	* parseTweetsAll --> extrae las palabras y sus ocurrencias de una lista de tweets
	* parseTweetsKeyword --> extrae las palabras y sus ocurrencias de una lista de tweets
				 que contienen la palabra clave dada
	* GetTweetsRange --> Saca todos los tweets en un rango de tiempo de una coleccion
	* GetTweetsAll --> Saca todos los tweets de una coleccion

v5
- Versión 4 + Añadidas las funciones de calcular intervalos de tiempo:
	* GetIntervalsLen --> Calcula intervalos de una longitud determinada en una franja de tiempo dada
	* GetIntervalsN --> Calculala N intervalos en una franja de tiempo dada

v6
- Version 5 + Añadidas funciones para obtener tweets de usuarios 'verifieds' y la lista de usuarios y numTweets
	* Creada libreria "trendings" con todas las funciones
	* GetTweetsVerified --> Saca todos los tweets de usuarios verificados
	* ListVerifiedUsers --> Devuelve un mapa con los usuarios verificados y el numero de tweets enviados

v7
- Version 6 + Añadida función de lectura de la BD
	* GetTweetsKeyword --> Saca todos los tweets y devuelve la lista de los 
				que contienen la palabra 'keyword'

v8
- Version 7 + Añadida función de más retweeteados y favoriteds a "profiles.go":
	* GetMostRetweetedAndFavorited --> Saca los 5 tweets más retweeteados y los 5
						con más favoritos de toda la lista

v9
- Version8 +: 	Modificada la conversion del texto a palabras
		Modificadas las funciones de "Parse" para que filtren el 'hashtag' y 'keyword'
		Añadidas gráficas de ejemplo al 'procesador.go' y al template de bootstrap
		Añadida una función para partir texto en "dobles palabras"

v10
- Version9 +: 	Añadidas las dobles palabras a la función ParseTweetsAll, ParseTweetsKeyword
		Introducida función de cálculo de subtrendings (considerando las palabras y dobles palabras)
		"MergeTrendings"
v11
- Version10+: 	Añadida la presentación ORDENADA de los tweets de usuarios verificados
		Corregido un bug en la funcion de filtrar dobles palabras por hashtag DoubleWordsFilter
		
================================
- PARA ABAJO INCOMPLETO -
================================
LISTA DE FUNCIONES PRINCIPALES:
================================

func parseTweetsAll (tweets []twittertypes.Tweet, hashtag string) wordsList 

func parseTweetsKeyword (tweets []twittertypes.Tweet, keyword string, hashtag string) wordsList 

func GetTweetsRange (colec mgo.Collection, start, end time.Time) []twittertypes.Tweet 

func GetTweetsAll (colec mgo.Collection) []twittertypes.Tweet 

func ListVerifiedUsers (tweets []twittertypes.Tweet) map[int64]VerifiedUser

func GetTweetsVerified (colec mgo.Collection) []twittertypes.Tweet

func GetIntervalsN (start, end time.Time, n int64) []Interval

func GetIntervalsLen (start, end time.Time, length time.Duration) []Interval


================================
LISTA DE FUNCIONES AUXILIARES:
================================
func AddWord(m map[string]int, word string) 

func TextToWords (t string) []string 

func ContainsWord (list []string, word string) bool 

func IsBlackListed (blacklist []string, word string) bool 



