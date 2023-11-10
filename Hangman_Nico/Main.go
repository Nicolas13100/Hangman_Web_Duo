package main

import (
	Hangman "Hangman/rsc"
	"fmt"
	"os"
)

func main() { // Menu d'entrer

	fmt.Println("Bienvenue sur notre jeu de Pendu")
	fmt.Println("Que voulez vous faire ? (1.jouer / 0. Partir)")
	for {
		var choice int
		fmt.Scan(&choice)
		switch choice {
		case 1:
			Hangman.Init()
			Hangman.Menu()
		case 0:
			os.Exit(0)
		default:
			fmt.Println("Choix Invalide")
			continue
		}
	}
}
