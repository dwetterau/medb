package main

import (
	"fmt"
	"syscall"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {

	fmt.Println("Enter password:")
	passwordBytes, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		panic(err)
	}

	hash, err := bcrypt.GenerateFromPassword(passwordBytes, 15)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(hash))
}
