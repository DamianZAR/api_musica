package main

import (
	"encoding/json"
	//"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

func main() {

	//Renderizado de la plantilla html
	http.HandleFunc("/", Index)
	http.HandleFunc("/consulta", consulta)

	//Creación del servidor
	fmt.Println("El servidor está corriendo en el puerto 8080.")
	fmt.Println("Run server: http://localhost:8080")
	http.ListenAndServe("localhost: 8080", nil)

}

func Index(w http.ResponseWriter, r *http.Request) {
	template, _ := template.ParseFiles("templates/index.html")
	template.Execute(w, nil)
}

type musica struct {
	Id       int
	Song     string
	Album    string
	Artist   string
	Duration string
	Artwork  string
	Price    string
	Origin   string
	Url      string
}

func consulta(w http.ResponseWriter, r *http.Request) {
	template, _ := template.ParseFiles("templates/index.html")

	song := r.FormValue("cancion")
	album := r.FormValue("album")
	artist := r.FormValue("artista")

	//Descarga y lectura de la respuesta json
	//url_api :=
	descargaApple(song, album, artist)
	sl := read_json()

	var id int
	var duration, artwork, price, origin, album_f, artist_f, song_f string
	album = strings.ToLower(album)
	artist = strings.ToLower(artist)
	for i := range sl {
		m, _ := sl[i].(map[string]interface{})

		album_f, _ = m["collectionName"].(string)
		artist_f, _ = m["artistName"].(string)

		artistf := strings.ToLower(artist_f)
		albumf := strings.ToLower(album_f)
		albumSP := strings.Split(albumf, " (")[0]

		if artistf == artist && albumSP == album {
			idl, _ := m["trackId"].(float64)
			id = int(idl)
			timel, _ := m["trackTimeMillis"].(float64)
			time := int(timel)
			time = time / 1000
			min := time / 60
			seg := time % 60
			duration = fmt.Sprintf("%02d:%02d", min, seg)
			artworkl, _ := m["artworkUrl100"].(string)
			artwork = string(artworkl)
			costl, _ := m["trackPrice"].(float64)
			cost := float64(costl)
			currencyl, _ := m["currency"].(string)
			currency := string(currencyl)
			price = fmt.Sprintf("$%.2f %s", cost, currency)

			song_f, _ = m["trackName"].(string)
			break
		}
	}

	//Extracción de la url de la letra de la canción
	url_lyr := lyrics(song, artist)

	cancion := musica{id, song_f, album_f, artist_f, duration, artwork, price, origin, url_lyr}

	template.Execute(w, cancion)
}

func descargaApple(song, album, artist string) string {
	//Extracción de la data desde la API de apple
	cancion_mod := strings.ReplaceAll(song, " ", "+")
	cancion_mod = strings.ToLower(cancion_mod)

	root := "https://itunes.apple.com/search?term="
	middleroot := "&media=music&entity="
	endroot := "song&attribute=songTerm&limit=150"

	//Descarga de la respuesta de la API de apple
	url_api := root + cancion_mod + middleroot + endroot
	response, err := http.Get(url_api)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	content, _ := io.ReadAll(response.Body)

	file := "json_apple.txt"
	err = ioutil.WriteFile(file, content, 0644)
	if err != nil {
		panic(err)
	}

	return url_api
}

func read_json() []interface{} {
	// Abrir el archivo en modo lectura
	file, err := os.Open("json_apple.txt")
	if err != nil {
		fmt.Println("Error al abrir el archivo:", err)
	}
	defer file.Close()

	// Decodificar el archivo JSON en un map dinámico
	var data map[string]interface{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		fmt.Println("Error al decodificar el archivo JSON:", err)
	}

	json_data := data["results"]

	var sl []interface{}

	if slice, ok := json_data.([]interface{}); ok {
		sl = slice
	} else {
		fmt.Println("No se puede convertir a un slice de interfaces")
	}

	return sl

}

func lyrics(song, artist string) string {

	artist_mod := strings.ReplaceAll(artist, " ", "+")
	song_mod := strings.ReplaceAll(song, " ", "+")
	root := "http://api.chartlyrics.com/apiv1.asmx/SearchLyric?artist="
	url := root + artist_mod + "&song=" + song_mod

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error al hacer solicitud a ChartLyrics.", err)
	}
	defer response.Body.Close()

	xml_data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error en la respuesta", err)
	}
	xml_text := string(xml_data)

	aux1 := strings.Split(xml_text, "<SearchLyricResult>")[1]
	aux2 := strings.Split(aux1, "<SongUrl>")[1]
	lyrics := strings.Split(aux2, "</SongUrl>")[0]
	//fmt.Printf("%T, %v\n", lyrics, lyrics)

	return lyrics

}
