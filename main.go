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
	History []string
}

var templates = template.Must(template.ParseGlob("templates/*.html"))

// In-memory session storage (in production, use proper session management)
var sessionHistory = make(map[string][]string)

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/history", historyHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(r)
	history := sessionHistory[sessionID]

	data := PageData{
		History: history,
	}
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

	sessionID := getSessionID(r)

	// Add to history if not already present
	addToHistory(sessionID, query)

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

func historyHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := getSessionID(r)
	history := sessionHistory[sessionID]

	data := PageData{
		History: history,
	}

	templates.ExecuteTemplate(w, "history.html", data)
}

func getSessionID(r *http.Request) string {
	// Simple session ID based on IP and User-Agent (in production, use proper session management)
	return r.RemoteAddr + r.UserAgent()
}

func addToHistory(sessionID, word string) {
	history := sessionHistory[sessionID]

	// Check if word already exists in history
	for _, w := range history {
		if strings.EqualFold(w, word) {
			return // Don't add duplicates
		}
	}

	// Add to beginning of history (most recent first)
	history = append([]string{word}, history...)

	// Keep only last 10 searches
	if len(history) > 10 {
		history = history[:10]
	}

	sessionHistory[sessionID] = history
}
