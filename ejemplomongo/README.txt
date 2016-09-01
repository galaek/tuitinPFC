capturador.go --> 	Captura los tweets del Stream filtrando por la palabra/s dadas
					Guarda los tweets en la base de datos
					
procesador.go --> 	Lee los tweets de la base de datos
					Parte el texto de cada tweet en palabras
					Cuenta la ocurrencia de cada palabra
					Guarda los pares [palabra, numero de veces] en la base de datos
					
resultados.go --> 	Lee los pares de palabras de la base de datos
					Muestra los resultados
					(resultados.go con 'resultado4' muestra un ejemplo hecho con 10 o 12 tweets)