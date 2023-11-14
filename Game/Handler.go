package Hangman

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

var (
	wordToGuess         string
	currentState        []string
	incorrectGuesses    []string
	playerName          string
	started             bool
	guessedLetters      = make(map[string]bool)
	incorrectGuessCount int
	difficulty          string
	invalidguess        string
)

func RUN() {
	// Set up your other handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/guess", guessHandler)
	http.HandleFunc("/lost", lostHandler)
	http.HandleFunc("/win", winHandler)
	http.HandleFunc("/restart", restartHandler)

	// Serve static files from the "static" directory
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Print statement indicating server is running
	fmt.Println("Server is running on :8080 http://localhost:8080")

	// Start the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if !started {
		renderTemplate(w, "start", nil)
		return
	}

	// Check if all letters have been guessed
	if !strings.Contains(strings.Join(currentState, ""), "_") {
		// If all letters have been guessed, redirect to the win page
		http.Redirect(w, r, "/win", http.StatusSeeOther)
		return
	}

	// Check if the player has reached the maximum number of incorrect guesses
	if incorrectGuessCount >= 8 {
		// If yes, redirect the player to the lost page
		http.Redirect(w, r, "/lost", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{"join": join}).ParseFiles("Template/template.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Started             bool
		PlayerName          string
		CurrentState        []string
		IncorrectGuesses    []string
		IncorrectGuessCount int
		Difficulty          string
		Invalidguess        string
	}{
		Started:             started,
		PlayerName:          playerName,
		CurrentState:        getCurrentState(),
		IncorrectGuesses:    incorrectGuesses,
		IncorrectGuessCount: incorrectGuessCount,
		Difficulty:          difficulty,
		Invalidguess:        invalidguess,
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
	difficulty = r.Form.Get("difficulty")
	// Do something with the difficulty
	// Load the word list based on difficulty
	wordList, err := loadWordList(difficulty)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Randomly select a word from the list
	wordToGuess = selectRandomWord(wordList)
	started = true
	resetCurrentState()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	guess := strings.ToUpper(r.Form.Get("guess"))

	if len(guess) >= 1 && !isLetter(guess) {
		invalidguess = "Invalid guess"
		http.Redirect(w, r, "/", http.StatusMovedPermanently)
		return
	} else {
		invalidguess = ""
	}

	if len(guess) == 1 {
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
			incorrectGuessCount++
		}
	} else {
		if guess == wordToGuess {
			http.Redirect(w, r, "/win", http.StatusSeeOther)
		} else {
			incorrectGuessCount += 2
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func lostHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		WordToGuess string
		// other fields...
	}{
		WordToGuess: wordToGuess,
		// other field values...
	}
	renderTemplate(w, "lost", data)
}

func winHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		PlayerName string
		// other fields...
		WordToGuess string
	}{
		PlayerName: playerName,
		// other field values...
		WordToGuess: wordToGuess,
	}
	renderTemplate(w, "win", data)
}

// Add a restartHandler to reset the game
func restartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	resetGame()
	resetCurrentState()

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
