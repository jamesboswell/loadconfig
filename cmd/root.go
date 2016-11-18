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

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	// RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.loadconfig.yaml)")
	// RootCmd.Flags().StringVar(&router, "router", "", "router name")
	RootCmd.PersistentFlags().StringVarP(&router, "router", "r", "", "Router hostname or IP address")
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "junos configuration commands file")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().BoolVarP(&Config.debug, "debug", "d", false, "Enable debugging")

	// if err != nil {
	// 	fmt.Printf("ERROR:: %s\n\n", err)
	// 	RootCmd.Help()
	// }

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

	// if cfgFile != "" { // enable ability to specify config file via flag
	// 	viper.SetConfigFile(cfgFile)
	// }
	//
	// viper.SetConfigName(".loadconfig") // name of config file (without extension)
	// viper.AddConfigPath("$HOME")       // adding home directory as first search path
	// viper.AutomaticEnv()               // read in environment variables that match
	//
	// // If a config file is found, read it in.
	// if err := viper.ReadInConfig(); err == nil {
	// 	fmt.Println("Using config file:", viper.ConfigFileUsed())
	// }
}

func checkFlags() error {
	err := ""
	if RootCmd.Flag("router").Changed != true {
		// logrus.Errorf("%s", "Router not specified")
		err = "Router not specified"
		// return errors.New(err)
	}
	if RootCmd.Flag("config").Changed != true {
		// logrus.Errorf("%s", "Configuration commands file not specified")
		err = "Configuration file not specified"
		// return errors.New(err)
	}
	if err != "" {
		return errors.New(err)
	}
	return nil
}
