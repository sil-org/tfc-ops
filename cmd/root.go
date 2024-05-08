// Copyright © 2018-2022 SIL International
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

	"github.com/silinternational/tfc-ops/v3/lib"
)

const requiredPrefix = "required - "

var (
	cfgFile      string
	organization string
	readOnlyMode bool
	errLog       *log.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:              "tfc-ops",
	Short:            "Terraform Cloud operations",
	Long:             `Perform TF Cloud operations, e.g. clone a workspace or manage variables within a workspace`,
	PersistentPreRun: initRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	errLog = log.New(os.Stderr, "", 0)
}

func initRoot(cmd *cobra.Command, args []string) {
	// Get Tokens from env vars
	atlasToken := os.Getenv("ATLAS_TOKEN")
	if atlasToken == "" {
		errLog.Fatalln("Error: Environment variable for ATLAS_TOKEN is required to execute plan and migration")
	}
	lib.SetToken(atlasToken)

	debugStr := os.Getenv("TFC_OPS_DEBUG")
	if debugStr == "TRUE" || debugStr == "true" {
		lib.EnableDebug()
	}

	if readOnlyMode {
		lib.EnableReadOnlyMode()
	}

	if err := lib.NewClient(""); err != nil {
		errLog.Fatalln(err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".tfc-ops" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".tfc-ops")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func addGlobalFlags(command *cobra.Command) {
	command.PersistentFlags().BoolVarP(&readOnlyMode, "read-only-mode", "r", false,
		`read-only mode (e.g. "-r")`,
	)

	command.PersistentFlags().StringVarP(&organization, "organization",
		"o", "", requiredPrefix+"Name of Terraform Cloud Organization")
	if err := command.MarkPersistentFlagRequired("organization"); err != nil {
		panic("MarkPersistentFlagRequired failed with error " + err.Error())
	}
}
