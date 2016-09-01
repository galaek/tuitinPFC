###En esta carpeta encontramos:
- capturador.go:
	* Configurando el hashtag (o hashtags) que queremos trackear y el nombre de la base de datos, captura los tweets de la API Stream de Twitter.

- procesador.go: 
	* Lee los tweets de una base de datos dada y los procesa de acuerdo a unos criterios definidos. Está configurado para lanzar un servidor en local y mostrar los resultados.

- web/: 
	* Contiene los archivos con contenido HTML, CSS y otros recursos web utilizados para la presentación de los resultados.