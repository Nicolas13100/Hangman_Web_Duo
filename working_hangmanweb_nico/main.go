package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

var (
	wordToGuess      = "hangman"
	currentState     = make([]string, len(wordToGuess))
	incorrectGuesses []string
	playerName       string
	started          bool
	guessedLetters   = make(map[string]bool)
)

func init() {
	// Initialize the current state with underscores
	for i := range currentState {
		currentState[i] = "_"
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/guess", guessHandler)

	fmt.Println("Server is running on :8080")
	http.ListenAndServe(":8080", nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if !started {
		renderTemplate(w, "start", nil)
		return
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{"join": join}).ParseFiles("template.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Started          bool
		PlayerName       string
		CurrentState     []string
		IncorrectGuesses []string
	}{
		Started:          started,
		PlayerName:       playerName,
		CurrentState:     currentState,
		IncorrectGuesses: incorrectGuesses,
	}

	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	resetGame()

	r.ParseForm()
	playerName = r.Form.Get("name")
	//difficulty := r.Form.Get("difficulty")
	// Do something with the difficulty

	started = true
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Assuming you want to get the player's name from a form field named "name"
	playerName = r.Form.Get("name")

	r.ParseForm()
	guess := strings.ToLower(r.Form.Get("guess"))

	if len(guess) != 1 || !isLetter(guess) {
		http.Error(w, "Invalid Guess", http.StatusBadRequest)
		return
	}

	if guessedLetters[guess] {
		// If the letter has already been guessed, do nothing
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	guessedLetters[guess] = true

	if strings.Contains(wordToGuess, guess) {
		updateState(guess)
	} else {
		incorrectGuesses = append(incorrectGuesses, guess)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func resetGame() {
	// Reset game-related variables
	currentState = make([]string, len(wordToGuess))
	incorrectGuesses = nil
	started = false
	guessedLetters = make(map[string]bool)

	// Initialize the current state with underscores
	for i := range currentState {
		currentState[i] = "_"
	}
}

func updateState(guess string) {
	for i, char := range wordToGuess {
		if string(char) == guess {
			currentState[i] = guess
		}
	}
}

func isLetter(s string) bool {
	return len(s) == 1 && ((s >= "a" && s <= "z") || (s >= "A" && s <= "Z"))
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{"join": join}).ParseFiles(tmplName + ".html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, tmplName, data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Template function to join slices
func join(s []string, sep string) string {
	return strings.Join(s, sep)
}
