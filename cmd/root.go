// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	AccessType   string
	OutputDir    string
	ProfilesFile string
	TagFile      string

	LogFile *os.File
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "aws-go-tool",
	Short: "aws-go-tool is an interface to use with aws accounts",
	Long: `The tool is designed around reporting and interacting with multiple aws accounts.
There are some parts of the tool that are just for single accounts as well.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer LogFile.Close()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.aws-go-tool.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	//TODO add flag checks to ensure the required flags are set
	RootCmd.PersistentFlags().StringVarP(&AccessType, "accessType", "a", "", "either assume, profile, instance, or instanceassume")
	RootCmd.PersistentFlags().StringVarP(&ProfilesFile, "profilesFile", "p", "", "file with list of account profiles")
	RootCmd.PersistentFlags().StringVarP(&TagFile, "tagFile", "g", "", "file with list of tags to add to output")
	RootCmd.PersistentFlags().StringVarP(&OutputDir, "outputDir", "o", "", "directory for script output")

	//Create output directory
	//utils.Dir("output")

	//Setup Log file
	//Close it in func Execute()
	LogFile, err := os.OpenFile("aws-go-tool.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	log.SetOutput(LogFile)
	log.Println("========== STARTING NEW RUN ==========")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".aws-go-tool") // name of config file (without extension)
	viper.AddConfigPath("$HOME")        // adding home directory as first search path
	viper.AutomaticEnv()                // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
