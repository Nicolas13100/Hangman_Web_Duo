package Hangman

import (
	cli "Hangman/Game/CLI"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
)

var (
	wordToGuess         string
	currentState        []string
	incorrectGuesses    []string
	playerName          string
	started             bool
	logged              bool
	lost                bool
	win                 bool
	guessedLetters      = make(map[string]bool)
	incorrectGuessCount int
	difficulty          string
	language            string
	invalidguess        string
	points              int
	score               int
	users               = make(map[string]User) // Map to store users
	username            string
	password            string
	globalData          = make(map[string]UserData)
	mutex               sync.Mutex
)

// User struct to represent user information
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserData structure for individual user data
type UserData struct {
	BestScore  int `json:"best_score"`
	TotalScore int `json:"total_score"`
}

type GlobalData map[string]UserData

func RUN() {
	// Load users from a file on startup
	if err := loadUsersFromFile("users.json"); err != nil {
		panic(err)
	}
	// Initialize global data
	globalData = make(GlobalData)

	// Set up your other handlers
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/start", startHandler)
	http.HandleFunc("/guess", guessHandler)
	http.HandleFunc("/lost", lostHandler)
	http.HandleFunc("/win", winHandler)
	http.HandleFunc("/restart", restartHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/confirmRegister", confirmRegisterHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/successLogin", successLoginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/scoreboard", scoreboardHandler)
	http.HandleFunc("/globalscoreboard", globalscoreboardHandler)
	http.HandleFunc("/gestion", gestionHandler)
	http.HandleFunc("/changeLogin", changeLoginHandler)

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
	if !logged {
		renderTemplate(w, "Login", nil)
		return
	}

	if !started {
		autoNaming := struct {
			Name string
		}{
			Name: username,
		}
		renderTemplate(w, "dashboard", autoNaming)
		return
	}
	// Check if all letters have been guessed
	if cli.Hangmanwin(currentState, wordToGuess) {
		win = true
		// If all letters have been guessed, redirect to the win page
		http.Redirect(w, r, "/win", http.StatusSeeOther)
		return
	}

	// Check if the player has reached the maximum number of incorrect guesses
	if cli.HangmanLost(incorrectGuessCount) {
		lost = true
		// If yes, redirect the player to the lost page
		http.Redirect(w, r, "/lost", http.StatusSeeOther)
		return
	}

	tmpl, err := template.New("index").Funcs(template.FuncMap{"join": join}).ParseFiles("Template/index.html")
	if err != nil {
		fmt.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	lefts := 8 - incorrectGuessCount
	data := struct {
		Logged              bool
		Started             bool
		PlayerName          string
		CurrentState        []string
		IncorrectGuesses    []string
		IncorrectGuessCount int
		Difficulty          string
		Invalidguess        string
		Points              int
		Score               int
		TriesLeft           int
	}{
		Logged:              logged,
		Started:             started,
		PlayerName:          playerName,
		CurrentState:        getCurrentState(),
		IncorrectGuesses:    incorrectGuesses,
		IncorrectGuessCount: incorrectGuessCount,
		Difficulty:          difficulty,
		Invalidguess:        invalidguess,
		Points:              points,
		Score:               score,
		TriesLeft:           lefts,
	}
	err = tmpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		fmt.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !started {
		resetGame()
		started = true
		autoNaming := struct {
			Name string
		}{
			Name: username,
		}
		renderTemplate(w, "start", autoNaming)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	playerName = r.Form.Get("name")
	difficulty = r.Form.Get("difficulty")
	language = r.Form.Get("language")
	// Do something with the difficulty
	// Load the word list based on difficulty
	wordList, err := cli.LoadWordList(language, difficulty)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Randomly select a word from the list
	wordToGuess = cli.SelectRandomWord(wordList)

	logged = true
	resetCurrentState()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !started {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}
	if win {
		http.Redirect(w, r, "/win", http.StatusSeeOther)
		return
	}
	if lost {
		http.Redirect(w, r, "/lost", http.StatusSeeOther)
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	guess := strings.ToUpper(r.Form.Get("guess"))

	if len(guess) >= 1 && !cli.IsLetter(guess) {
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
			calculateScoreWin()
		} else {
			incorrectGuesses = append(incorrectGuesses, guess)
			incorrectGuessCount++
			calculateScoreLose()
		}
	} else {
		fmt.Print(guess)
		if guess == wordToGuess && !lost {
			calculateScoreFinal()
			win = true
			http.Redirect(w, r, "/win", http.StatusSeeOther)
		} else if guess == "INSTANTWIN" {
			fmt.Println("ok")
			renderTemplate(w, "winning", nil)
		} else {
			incorrectGuessCount += 2
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func lostHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !started {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}
	data := struct {
		WordToGuess string
		Difficulty  string
		// other fields...
	}{
		WordToGuess: wordToGuess,
		Difficulty:  difficulty,
		// other field values...
	}
	renderTemplate(w, "lost", data)
}

func winHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !started {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}
	err0 := UpdateAndSaveGlobalData(username, score)
	if err0 != nil {
		http.Error(w, "Failed to save Global data", http.StatusInternalServerError)
		// Log the error or handle it appropriately
		log.Println("Error saving Global data:", err0)
		return
	}
	err := SaveUserData()
	if err != nil {
		http.Error(w, "Failed to save user data", http.StatusInternalServerError)
		// Log the error or handle it appropriately
		log.Println("Error saving user data:", err)
		return
	}

	data := struct {
		PlayerName string
		// other fields...
		WordToGuess string
		Score       int
		Difficulty  string
	}{
		PlayerName: playerName,
		// other field values...
		WordToGuess: wordToGuess,
		Score:       score,
		Difficulty:  difficulty,
	}
	renderTemplate(w, "win", data)
}

// Add a restartHandler to reset the game
func restartHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if !started {
		http.Redirect(w, r, "/index", http.StatusSeeOther)
		return
	}

	if !win && !lost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	resetGame()
	resetCurrentState()

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "Register", nil)
}

func confirmRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Check if username already exists
	if _, exists := users[username]; exists {
		http.Error(w, "Username already exists", http.StatusConflict)
		return
	}

	hashedPassword := hashPassword(password)
	users[username] = User{Username: username, Password: hashedPassword}

	// Save users to a file
	if err := saveUsersToFile("users.json"); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Load users from a file on startup
	if err := loadUsersFromFile("users.json"); err != nil {
		panic(err)
	}
	renderTemplate(w, "Login", nil)
}

func successLoginHandler(w http.ResponseWriter, r *http.Request) {
	logged = true
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	username = r.FormValue("username")
	password = r.FormValue("password")

	user, exists := users[username]
	if !exists || !checkPasswordHash(password, user.Password) {
		fmt.Println("Invalid username or password")
		return
	}

	// Successfully logged in
	// Handle further operations (e.g., setting session, redirecting, etc.)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	resetUserValue()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	autoData := struct {
		Name string
	}{
		Name: username,
	}
	renderTemplate(w, "dashboard", autoData)
}

func scoreboardHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	BestScore, Score, err := extractVariablesFromJSONFile()
	if err != nil {
		fmt.Print("extract not working")
	}
	data := struct {
		PlayerName string
		BestScore  int
		TotalScore int
	}{
		PlayerName: username,
		BestScore:  BestScore,
		TotalScore: Score,
	}
	renderTemplate(w, "scoreboard", data)
}

func globalscoreboardHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	userDataList, err := globalextractVariablesFromJSONFile()
	if err != nil {
		fmt.Println("Error extracting data:", err)
		// Handle the error (e.g., return an error response to the client)
		return
	}

	// Create a slice to hold data for all users
	var allUsersData []struct {
		PlayerName string
		BestScore  int
		TotalScore int
	}

	// Process data for each user
	for playerName, userData := range userDataList {
		userDataEntry := struct {
			PlayerName string
			BestScore  int
			TotalScore int
		}{
			PlayerName: playerName,
			BestScore:  userData.BestScore,
			TotalScore: userData.TotalScore,
		}

		// Append data for each user to the slice
		allUsersData = append(allUsersData, userDataEntry)
	}
	renderTemplate(w, "globalscoreboard", allUsersData)
}

func gestionHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	data := struct {
		PlayerName string
	}{
		PlayerName: username,
	}
	renderTemplate(w, "gestion", data)
}

func changeLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !logged {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	oldpassword := r.FormValue("oldpassword")
	newpassword := r.FormValue("newpassword")
	err := updateUserCredentials(username, oldpassword, newpassword)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Password updated successfully.")
	resetUserValue()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
