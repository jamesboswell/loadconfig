// Copyright © 2016 James Boswell <boswell.jim@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ProgramName -
var ProgramName = "loadconfig"

// Version -
var Version = "1.0.0"

// Config is the global config
var Config struct {
	router string
	file   string
	debug  bool
}

var (
	router     string
	configFile string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "loadconfig",
	Short: "A Junos configuration loader",
	Long:  `loadconfig loads Juniper Junos configuration commands via NETCONF`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Flags
	RootCmd.PersistentFlags().StringVarP(&router, "router", "r", "", "Router hostname or IP address")
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "junos configuration commands file")
	// Debug flag
	RootCmd.PersistentFlags().BoolVarP(&Config.debug, "debug", "d", false, "Enable debugging")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if RootCmd.Flag("router").Changed != true {
		RootCmd.Help()
		fmt.Println()
		logrus.Errorf("%s", "--router or -r parameter not provided")
		fmt.Println()
		os.Exit(1)
	}
	if RootCmd.Flag("config").Changed != true {
		RootCmd.Help()
		fmt.Println()
		logrus.Errorf("%s", "--config or -c parameter not provided")
		fmt.Println()
		os.Exit(1)
	}

	Config.router = router
	Config.file = configFile

}

func checkFlags() error {
	err := ""
	if RootCmd.Flag("router").Changed != true {
		err = "Router not specified"
	}
	if RootCmd.Flag("config").Changed != true {
		err = "Configuration file not specified"
	}
	if err != "" {
		return errors.New(err)
	}
	return nil
}
