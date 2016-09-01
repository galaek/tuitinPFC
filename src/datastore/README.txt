- Editar en 'timelines.go' las variables InteresesDB y PersonalidadesDB si interesa cambiar de BD.
	* Actualmente "intereses3" y  "personalidades3" contienen la BD en uso

- analizador.go:
	* Lee una timeline de Twitter y la analiza usando la BD existente de palabras ofreciendo sus resultados

- generador.go: 
	* Borra la BD de palabras existente. 
	* Lee todas las timelines de las encuestas. (esto le cuesta +30mins)
	* Procesa las timelines y guarda la lista de palabras en la BD.

- carga.go:
	* Contiene un main de prueba para 'timelines.go'. De ahi se puede copiar código.
	* Mejor no ejecutar esto...

- CREDENTIALS:
	* Es usado por 'profiles.go' para leer las credenciales de twitter