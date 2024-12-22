package terminal

import (
	"bufio"
	"fmt"
	"golang.org/x/term"
	"os"
	"strings"
	"syscall"
)

type Term struct{}

type Terminal interface {
	PromptPassword(promptMsg string) (string, error)
}

// PromptPassword prompts user to enter password and then returns it
func (c *Term) PromptPassword(promptMsg string) (string, error) {
	fmt.Print(promptMsg)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

// PromptPassword prompts user to enter password and then returns it
func PromptPassword(promptMsg string) (string, error) {
	fmt.Print(promptMsg)
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	return strings.TrimSpace(string(bytePassword)), nil
}

// InputPrompt receives a string value using the label
func InputPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}

// AuthPrompt implements the Telegram auth flow interface
type AuthPrompt struct {
	PhoneNumber string
}

// Code prompts the user to enter the code sent by Telegram
func (a AuthPrompt) Code() (string, error) {
	return InputPrompt("Enter the code you received: "), nil
}

// Password prompts the user to enter their 2FA password
func (a AuthPrompt) Password() (string, error) {
	return PromptPassword("Enter your password (if applicable): ")
}

// Phone returns the phone number provided
func (a AuthPrompt) Phone() string {
	return a.PhoneNumber
}
