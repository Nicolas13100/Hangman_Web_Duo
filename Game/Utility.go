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
	wordToGuess = ""
	currentState = []string{}
	incorrectGuesses = []string{}
	playerName = ""
	started = false
	guessedLetters = make(map[string]bool)
	incorrectGuessCount = 0
	difficulty = ""
	invalidguess = ""
	points = 0
	score = 0
}

func resetCurrentState() {
	// Initialize the current state with underscores
	currentState = make([]string, len(wordToGuess))
	fmt.Println(len(currentState))
	fmt.Println(len(wordToGuess))
	for i := range currentState {
		currentState[i] = "_"
	}
	numIterations := int(float64(len(wordToGuess)) * 0.35)
	for i := 0; i < numIterations; i++ {
		randIndex := rand.Intn(len(wordToGuess))
		letter := string(wordToGuess[randIndex])
		updateState(letter)
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
	if len(s) == 0 {
		return false // Return false for an empty string
	}

	for _, char := range s {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
			return false // If any character is not a letter, return false
		}
	}

	return true // If all characters are letters, return true
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

func calculateScoreWin() {
	switch difficulty {
	case "Dictionnaire":
		score += 10
	case "Facile":
		score += 5
	case "Moyen":
		score += 15
	case "Diffile":
		score += 20
	case "Halloween":
		score += 10
	default:
		fmt.Println("Can't find difficulty")
	}
}

func calculateScoreLose() {
	switch difficulty {
	case "Dictionnaire":
		score -= 5
	case "Facile":
		score -= 1
	case "Moyen":
		score -= 7
	case "Diffile":
		score -= 10
	case "Halloween":
		score -= 5
	default:
		fmt.Println("Can't find difficulty")
	}
}

func calculateScoreFinal() {

	score = score + 5*countUnderscores(currentState)
}

func countUnderscores(arr []string) int {
	count := 0

	for _, val := range arr {
		if val == "_" {
			count++
		}
	}

	return count
}
