## En esta carpeta encontramos:
- analizador.go:
	* Lee una timeline de Twitter y la analiza usando la BD existente de palabras y nos da sus resultados.

- generador.go: 
	* Borra la BD de palabras existente. 
	* Lee todas las timelines de las encuestas.
	* Procesa las timelines y guarda la lista de palabras en la base de conocimiento.

- verificador.go:
	* Programa creado para la verificación del criterio de análisis de perfiles.

- CREDENTIALS:
	* Es usado por 'profiles.go' para leer las credenciales de Twitter
	