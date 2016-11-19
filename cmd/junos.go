package cmd

import (
	"fmt"
	"os"

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
		// start the NETCONF session
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
	fmt.Printf("\n\n")

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

	c.Println("Gathering system facts ...")
	hr()
	// Print system info
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

	fmt.Printf("Config diff (show | compare)::\n")
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
			g.Println("Final commit is SUCCESSFUL!")
		}

	} else {
		r.Println("Commit not finalized - initiate rollback 1")
		if err = s.RollbackConfig(1); err != nil {
			l.Error("Rollback failed", err)
		}
	}

	hr()
	y.Println("Config changes commited, config unlocked")
	g.Println("Closing connection, Goodbye!")
	hr()
	s.Close()
	fmt.Println("Connection closed to: ", Config.router)
}
