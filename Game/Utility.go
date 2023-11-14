package Hangman

import (
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func resetGame() {
	// Reset game-related variables
	incorrectGuessCount = 0
	incorrectGuesses = nil
	started = false
	guessedLetters = make(map[string]bool)
}

func resetCurrentState() {
	// Initialize the current state with underscores
	currentState = make([]string, len(wordToGuess))
	fmt.Println(len(currentState))
	for i := range currentState {
		currentState[i] = "_"
	}
}

func updateState(guess string) {
	wordRunes := []rune(wordToGuess)
	guessRune := []rune(guess)[0]

	for i, char := range wordRunes {
		if char == guessRune {
			fmt.Printf("Updating currentState[%d] to %s\n", i, guess)
			currentState[i] = guess
		}
	}
}

func getCurrentState() []string {
	return currentState
}

func isLetter(s string) bool {
	return len(s) == 1 && ((s >= "a" && s <= "z") || (s >= "A" && s <= "Z"))
}

func renderTemplate(w http.ResponseWriter, tmplName string, data interface{}) {
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{"join": join}).ParseFiles("Template/" + tmplName + ".html")
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

// Helper function to load word list based on difficulty
func loadWordList(difficulty string) ([]string, error) {
	// Construct the file path based on the difficulty level
	filePath := fmt.Sprintf("Librairie/%s.txt", difficulty)

	// Read the content of the file
	content, err := os.ReadFile(filePath)
	fmt.Println(err)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	// Split the content into lines to get individual words
	wordList := strings.Split(string(content), "\n")

	// Filter out empty lines
	var filteredWordList []string
	for _, word := range wordList {
		// Skip empty lines
		if word != "" {
			filteredWordList = append(filteredWordList, word)
		}
	}

	if len(filteredWordList) == 0 {
		return nil, fmt.Errorf("no words found in the file")
	}

	return filteredWordList, nil
}

// Helper function to select a random word from the list
func selectRandomWord(wordList []string) string {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	return wordList[randGenerator.Intn(len(wordList))]
}
