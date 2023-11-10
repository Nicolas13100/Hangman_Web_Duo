package Hangman

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
)

var hangman []string

func Init() {
	// Read hangman ASCII art from file
	data, err := os.ReadFile("AsciiArt/hangman.txt")
	if err != nil {
		fmt.Println("Error reading hangman file:", err)
		return
	}

	if len(data) == 0 {
		fmt.Println("Error: hangman file is empty")
		return
	}
	hangman = strings.Split(string(data), "\n")
}

func PrintHangman(incorrectGuesses int) { // Print the hangman to correct position
	if incorrectGuesses == 0 {
		fmt.Println(hangman[0])
	} else if incorrectGuesses <= len(hangman) {
		start := (incorrectGuesses - 1) * 7
		end := incorrectGuesses * 7

		if end > len(hangman) {
			end = len(hangman)
		}

		for i := start; i < end; i++ {
			fmt.Println(hangman[i])
		}
	} else {
		fmt.Println(hangman[len(hangman)-1])
	}
}

func Pendu(s string) { // Le jeu
	fmt.Println("C'est parti, a vous de jouer")
	word := s
	GuessedLetters := []string{}
	incorrectGuesses := 0
	result := false
	numIterations := int(float64(len(word)) * 0.2)
	for i := 0; i < numIterations; i++ {
		randIndex := rand.Intn(len(word))
		letter := string(word[randIndex])
		GuessedLetters = append(GuessedLetters, letter)
	}
	displayWord := ""
	for _, letter := range word { // Ecrit _ si letter non deviné pour l'initialisation
		if contains(GuessedLetters, string(letter)) {
			displayWord += string(letter) + " "
		} else {
			displayWord += "_ "
		}
	}

	fmt.Println(displayWord)

	for { // Prend la valeur user et la vérifie puis applique si ok
		var user_input string
		if len(GuessedLetters) > 0 {
			fmt.Printf("Lettre deja donné: %s\n", strings.Join(GuessedLetters, ", "))
		}
		fmt.Print("Entrer une lettre ou un mot: ")
		_, err := fmt.Scan(&user_input)

		if err != nil {
			fmt.Println("Erreur de saisie:", err)
			continue
		}

		if len(user_input) == 1 && strings.ContainsAny(user_input, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			result, GuessedLetters = checkLetter(word, GuessedLetters, strings.ToUpper(user_input))

			if !result {
				incorrectGuesses++
			}
		} else if len(user_input) == len(word) && strings.ContainsAny(user_input, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			USER := strings.ToUpper(user_input)
			if USER == word { // Test si l'utilsateur a tenté d'entrer un mot
				fmt.Printf("Bravo, vous avez trouvé le mot: %s\n", word)
				break
			} else {
				incorrectGuesses += 2
			}
		} else {
			fmt.Println("Entrer une lettre ou un mot valide SVP.")
			continue
		}

		PrintHangman(incorrectGuesses)

		displayWord := ""
		for _, letter := range word { // Ecrit _ par lettre manquante
			if contains(GuessedLetters, string(letter)) {
				displayWord += string(letter) + " "
			} else {
				displayWord += "_ "
			}
		}

		fmt.Println(displayWord)

		if strings.Join(strings.Fields(displayWord), "") == word { // Win condition for found with lettre by lettre
			fmt.Printf("Bravo, vous avez trouvé le mot: %s\n", word)
			break
		}

		if incorrectGuesses == 9 {
			fmt.Printf("Dommage, plus de tentative. Le mot était: %s\n", word)
			break
		}
	}
}

func contains(slice []string, item string) bool { // compare et renvoi si oui ou non (es'que a contien b)
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

func checkLetter(word string, GuessedLetters []string, letter string) (bool, []string) { // Verifie si la lettre a deja ete donné
	for _, guessedLetter := range GuessedLetters {
		if guessedLetter == letter {
			ClearTerminal()
			fmt.Printf("Vous avez déja essayer la lettre '%s'.\n", letter)
			return true, GuessedLetters
		}
	}

	if strings.Contains(word, letter) {
		GuessedLetters = append(GuessedLetters, letter)
		return true, GuessedLetters
	} else {
		GuessedLetters = append(GuessedLetters, letter)
		return false, GuessedLetters
	}
}
