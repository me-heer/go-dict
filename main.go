package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type DictionaryEntry struct {
	Word      string     `json:"word"`
	Phonetic  string     `json:"phonetic"`
	Phonetics []Phonetic `json:"phonetics"`
	Meanings  []Meaning  `json:"meanings"`
}

type Phonetic struct {
	Text  string `json:"text"`
	Audio string `json:"audio"`
}

type Meaning struct {
	PartOfSpeech string       `json:"partOfSpeech"`
	Definitions  []Definition `json:"definitions"`
}

type Definition struct {
	Definition string   `json:"definition"`
	Example    string   `json:"example"`
	Synonyms   []string `json:"synonyms"`
	Antonyms   []string `json:"antonyms"`
}

type PageData struct {
	Results []DictionaryEntry
	Error   string
	Query   string
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{}
	templates.ExecuteTemplate(w, "index.html", data)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := strings.TrimSpace(r.FormValue("word"))
	if query == "" {
		templates.ExecuteTemplate(w, "results.html", PageData{Error: "Please enter a word"})
		return
	}

	// Fetch from dictionary API
	apiURL := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", url.QueryEscape(query))
	resp, err := http.Get(apiURL)
	if err != nil {
		templates.ExecuteTemplate(w, "results.html", PageData{Error: "Failed to fetch dictionary data"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		templates.ExecuteTemplate(w, "results.html", PageData{Error: "Word not found. Please check your spelling and try again."})
		return
	}

	if resp.StatusCode != 200 {
		templates.ExecuteTemplate(w, "results.html", PageData{Error: "Failed to fetch dictionary data"})
		return
	}

	var results []DictionaryEntry
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		templates.ExecuteTemplate(w, "results.html", PageData{Error: "Failed to parse dictionary data"})
		return
	}

	data := PageData{
		Results: results,
		Query:   query,
	}

	templates.ExecuteTemplate(w, "results.html", data)
}
