package Hangman

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

func resetGame() {
	wordToGuess = ""
	currentState = []string{}
	incorrectGuesses = []string{}
	playerName = ""
	started = false
	lost = false
	win = false
	gameUpdated = false
	guessedLetters = make(map[string]bool)
	incorrectGuessCount = 0
	difficulty = ""
	invalidguess = ""
	points = 0
	score = 0
}

func ResetUserValue() {
	logged = false
	username = ""
	password = ""
	resetGame()
}

func resetCurrentState() {
	// Initialize the current state with underscores
	currentState = make([]string, len(wordToGuess))

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
			currentState[i] = guess
		}
	}
}

func getCurrentState() []string {
	return currentState
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
	case "Christmas":
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
		score -= 5
	case "Christmas":
		score -= 5
	default:
		fmt.Println("Can't find difficulty")
	}
}

func calculateScoreFinal() {

	score = score + 25*countUnderscores(currentState)
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

// Function to load users from a file for register func
func loadUsersFromFile(filename string) error {
	// Check if the file exists
	fileInfo, err := os.Stat(filename)
	if os.IsNotExist(err) {
		// Create an empty users.json file if it doesn't exist
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
	} else if err != nil {
		return err
	}

	// Check if the file is empty
	if fileInfo != nil && fileInfo.Size() == 0 {
		// File exists but is empty, so initialize users as an empty map
		users = make(map[string]User)
		return nil
	}

	// Load users from the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	// Check if the file contains valid JSON data
	if len(data) == 0 {
		// File is empty or doesn't contain valid JSON
		return nil
	}

	users = make(map[string]User)
	if err := json.Unmarshal(data, &users); err != nil {
		return err
	}

	return nil
}

// Function to save user data to a file
func SaveUserData() error {
	filename := username + ".json"
	var userData UserData

	// Open the file with read-write access or create if it doesn't exist
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decode existing data from the file
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	if stat.Size() != 0 {
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&userData); err != nil {
			return err
		}
	}

	// Update scores if the new score is better
	if score > userData.BestScore {
		userData.BestScore = score
	}
	userData.TotalScore += score // Add to the total score

	// Seek to the beginning of the file before writing
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// Truncate the file before writing to ensure no leftover data
	if err := file.Truncate(0); err != nil {
		return err
	}

	// Encode and write the updated data back to the file
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(userData); err != nil {
		return err
	}

	return nil
}

// Function to update global data and save it to a file
func UpdateAndSaveGlobalData(userID string, Score int) error {
	mutex.Lock()
	defer mutex.Unlock()

	if _, exists := globalData[userID]; !exists {
		globalData[userID] = UserData{}
	}

	user := globalData[userID]

	if Score > user.BestScore {
		user.BestScore = Score
	}
	user.TotalScore += Score

	globalData[userID] = user

	// Save global data to file
	return saveGlobalDataToFile()
}

// Function to save global data to a file
func saveGlobalDataToFile() error {
	filename := "global_data.json"

	data, err := json.Marshal(globalData)
	if err != nil {
		return err
	}

	// Write data to the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func extractVariablesFromJSONFile() (int, int, error) {
	filePath := username + ".json"
	// Read JSON file
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return 0, 0, err
	}

	// Define a struct to represent the JSON data
	type User struct {
		BestScore  int `json:"best_score"`
		TotalScore int `json:"total_score"`
	}

	// Unmarshal JSON data into the struct
	var user User
	err = json.Unmarshal(byteValue, &user)
	if err != nil {
		return 0, 0, err
	}

	return user.BestScore, user.TotalScore, nil
}

func globalextractVariablesFromJSONFile() (GlobalData, error) {
	filePath := "global_data.json"

	// Try to open JSON file
	file, err := os.Open(filePath)
	if err != nil {
		// If unable to open, create a new empty GlobalData and return
		if os.IsNotExist(err) {
			return GlobalData{}, nil
		}
		return GlobalData{}, err
	}
	defer file.Close()

	byteValue, err := io.ReadAll(file)
	if err != nil {
		return GlobalData{}, err
	}

	// Unmarshal JSON data into the struct
	var globalData GlobalData
	err = json.Unmarshal(byteValue, &globalData)
	if err != nil {
		return GlobalData{}, err
	}

	return globalData, nil
}

func updateUserCredentials(name, oldPassword, newPassword string) error {
	// Read the JSON file into memory
	raw, err := os.ReadFile("users.json")
	if err != nil {
		return err
	}

	// Define a struct that matches your JSON structure
	var data map[string]User // Map where keys are strings and values are User structs

	// Unmarshal the JSON into the defined struct
	if err := json.Unmarshal(raw, &data); err != nil {
		return err
	}

	// Check if the user exists in the map
	user, exists := data[name]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if !checkPasswordHash(oldPassword, user.Password) {
		return fmt.Errorf("incorrect password")
	}

	if newPassword != "" {
		// Update the password
		user.Password = hashPassword(newPassword)

		// Update the user in the map
		data[name] = user

		// Marshal the updated data back to JSON
		updatedJSON, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}

		// Write the updated JSON back to the file
		err = os.WriteFile("users.json", updatedJSON, 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

// Function to save users to a file for register func
func saveUsersToFile(filename string) error {
	data, err := json.Marshal(users)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		log.Println("Error writing updated user data:", err)
		return err
	}

	log.Println("User data successfully updated")
	return nil
}

func hashPassword(password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(password))
	hashedPassword := hasher.Sum(nil)
	return hex.EncodeToString(hashedPassword)
}

func checkPasswordHash(password, hash string) bool {
	hashedPassword := hashPassword(password)
	return hashedPassword == hash
}
