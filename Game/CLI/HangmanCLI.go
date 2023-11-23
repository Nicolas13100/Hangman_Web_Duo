package cli

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

func Hangmanwin(currentState []string, wordToGuess string) bool {
	// Check if all letters have been guessed
	if stringifyStringSlice(currentState) == wordToGuess && wordToGuess != "" {
		return true
	}
	return false
}

func stringifyStringSlice(strSlice []string) string {
	var result string
	for _, s := range strSlice {
		result += s
	}
	return result
}

func HangmanLost(incorrectGuessCount int) bool {
	return incorrectGuessCount >= 8
}

func SelectRandomWord(wordList []string) string {
	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)
	return wordList[randGenerator.Intn(len(wordList))]
}

func LoadWordList(language, difficulty string) ([]string, error) {
	// Construct the file path based on the difficulty level
	filePath := fmt.Sprintf("Librairie/%s/%s.txt", language, difficulty)

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
		word = strings.TrimSpace(word)
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

func IsLetter(s string) bool {
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
