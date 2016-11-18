package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/Bowery/prompt"
)

// promptUser prompts user for information
func promptUser(field string) (string, error) {

	value, err := prompt.Basic(field, true)
	if err != nil {
		return "", err
	}

	return value, err
}

// readConfig reads in config file (does not parse!)
func readConfig(file string) (string, error) {
	f, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Print(err)
	}

	return string(f), err
}

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
// Credit albrow : https://gist.github.com/albrow/5882501
func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

// posString returns the first index of element in slice.
// If slice does not contain element, returns -1.
func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

// hr prints a horizontal line i times
func hr(i ...int) {
	numLines := 1
	if i != nil {
		numLines = i[0]
	}
	for x := 0; x < numLines; {
		fmt.Printf("%s\n", strings.Repeat("--", 20))
		x++
	}
}
