package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/Bowery/prompt"
	"github.com/Sirupsen/logrus"
	"github.com/fatih/color"
	j "github.com/jamesboswell/go-junos"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var user string
var password string

var junosCmd = &cobra.Command{
	Use:   "junos",
	Short: "Loads Juniper Junos configurations",
	Long:  `Loads Juniper Junos configuration from --config file`,
	// Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Work your own magic here
		// fmt.Println("junos called")

		openSession()
	},
}

func init() {
	RootCmd.AddCommand(junosCmd)
	// Allow user and passowrd to be passed as flags
	junosCmd.PersistentFlags().StringVarP(&user, "user", "u", "", "Username")
	junosCmd.PersistentFlags().StringVarP(&password, "pass", "p", "", "Password")

}

// openSession starts the NETCONF session
func openSession() {

	l := logrus.New()
	l.Formatter = new(prefixed.TextFormatter)

	if Config.debug {
		l.Level = logrus.DebugLevel
		l.Debug(Config.router)
		l.Debug(Config.file)
	}

	//colors
	c := color.New(color.FgHiCyan).Add(color.Bold)
	y := color.New(color.FgHiYellow)
	g := color.New(color.FgHiGreen).Add(color.Bold)
	r := color.New(color.FgHiRed).Add(color.Bold)

	// Print info

	g.Printf("\n%s", ProgramName)
	fmt.Printf(" - %s\n", Version)
	hr()
	c.Printf("%s\n", "Starting session....")
	y.Printf("\t     router:  %-6s\n", Config.router)
	y.Printf("\tconfig file:  %-6s\n", Config.file)
	hr()

	// Prompt for credentials
	if user == "" {
		u, err := promptUser("Username: ")
		if err != nil {
			l.Fatalf("Username not provided")
		}
		user = u
	}
	if password == "" {
		password, _ = prompt.Password("Password: ")
	}

	// Start session
	s, err := j.NewSession(Config.router, user, password, l)
	defer s.Close()
	if err != nil {
		l.Errorf("NETCONF session error : %s", err)
	}

	gather := `
------------------------
Gathering system facts ...
------------------------`

	c.Println(gather)
	fmt.Printf("HOSTNAME :\t%-10s\n", s.Hostname)

	for _, p := range s.Platform {
		fmt.Printf("   MODEL :\t%-10s\n", p.Model)
		fmt.Printf("   JUNOS :\t%-10s\n", p.Version)
	}
	hr()

	// Lock
	y.Println("Locking candidate config")
	if err = s.Lock(); err != nil {
		l.Fatal("Unable to lock config", err)
	}

	//Make config change
	g.Printf("Loading config from %s ...\n", Config.file)

	setcommands, err := readConfig(Config.file)
	l.Debug(setcommands)

	if err != nil {
		l.Fatal("Unable to read config file")
	}

	err = s.Config(setcommands, "set", false)
	if err != nil {
		l.Fatal(err)
	}

	// Check config
	fmt.Println("Checking candidate config")
	if err = s.CommitCheck(); err != nil {
		l.Error("Commit check failed...")
		l.Error("Rolling back config")
		s.RollbackConfig(0)
		s.Unlock()
		s.Close()
		os.Exit(1)
	} else {
		fmt.Println("candidate config is OK!")
	}

	fmt.Printf("Config diff::\n")
	delta, err := s.ConfigDelta()
	if err != nil {
		l.Error("no difference in config?")
	}

	fmt.Println(delta)

	fmt.Println("Commit changes? [Y/n]")
	confirm := askForConfirmation()
	if confirm == true {
		fmt.Println("Commiting changes with 5 minute auto-rollback......")
		if err = s.CommitConfirm(5); err != nil {
			fmt.Println("Commit FAILED! -- init rollback 0")
			l.Error(err)
			s.RollbackConfig(0)
			s.Unlock()
			os.Exit(1)
		} else {
			g.Println("Commit SUCCESSFUL!")
			y.Println("configuration will automatically rollback in 5 mins")
		}

	} else {
		r.Println("Configuration not committed - initiate rollback 0")
		if err = s.RollbackConfig(0); err != nil {
			l.Error("Rollback failed", err)
		}
		hr()
		y.Println("Config changes rolled back, closing connection")
		if err = s.Unlock(); err != nil {
			l.Error("Unlock failed")
		}
		y.Println("Configuration Unlocked")
		s.Close()
		fmt.Println("Connection closed to: ", Config.router)
		os.Exit(0)
	}

	//Unlock
	if err = s.Unlock(); err != nil {
		l.Error("Unable to unlock canidate config: ", err)
	}

	// Confirm Commit
	hr()
	r.Println("Validate device status before continuning!")
	hr()
	fmt.Println("Confirm commit changes? [Y/n]")
	confirmFinal := askForConfirmation()
	if confirmFinal == true {
		fmt.Println("Commiting ......")
		if err = s.Commit(); err != nil {
			fmt.Println("Commit FAILED! -- init rollback 0")
			l.Error(err)
			s.RollbackConfig(0)
			s.Unlock()
			os.Exit(1)
		} else {
			fmt.Println("Final commit is SUCCESSFUL!")
		}

	} else {
		r.Println("Commit not finalized - initiate rollback 1")
		if err = s.RollbackConfig(1); err != nil {
			l.Error("Rollback failed", err)
		}
	}

	hr()
	y.Println("Config changes rolled back, closing connection")
	// if err = s.Unlock(); err != nil {
	// 	l.Error("Unlock failed")
	// }
	y.Println("Configuration Unlocked")
	hr()
	g.Println("Goodbye!")
	hr()
	s.Close()
	fmt.Println("Connection closed to: ", Config.router)
}

// promptUser prompts user for information
func promptUser(field string) (string, error) {

	value, err := prompt.Basic(field, true)
	if err != nil {
		return "", err
	}

	return value, err
}

func readConfig(file string) (string, error) {
	f, err := ioutil.ReadFile(file) // just pass the file name
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
