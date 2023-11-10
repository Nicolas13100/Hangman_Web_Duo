package Hangman

import (
	"os"
	"os/exec"
	"runtime"
)

func ClearTerminal() { // Néttoie la console
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
