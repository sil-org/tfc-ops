// Copyright © 2018-2022 SIL Global
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
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/silinternational/tfc-ops/v4/lib"
)

const requiredPrefix = "required - "

var (
	cfgFile      string
	organization string
	readOnlyMode bool
	debugMode    bool
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
	debugStr := os.Getenv("TFC_OPS_DEBUG")
	if debugStr == "TRUE" || debugStr == "true" {
		lib.EnableDebug()
		debugMode = true
	}

	if readOnlyMode {
		lib.EnableReadOnlyMode()
	}

	getToken()
}

type Credentials struct {
	Credentials struct {
		AppTerraformIo struct {
			Token string `json:"token"`
		} `json:"app.terraform.io"`
	} `json:"credentials"`
}

func getToken() {
	credentials, err := readTerraformCredentials()
	if err != nil {
		errLog.Fatalln("failed to get Terraform credentials:", err)
	}

	if credentials != nil {
		token := credentials.Credentials.AppTerraformIo.Token
		if token != "" {
			lib.SetToken(token)
			return
		}
	}

	// fall back to using ATLAS_TOKEN environment variable
	atlasToken := os.Getenv("ATLAS_TOKEN")
	if atlasToken != "" {
		lib.SetToken(atlasToken)
		return
	}

	errLog.Fatalln("no credentials found, use 'terraform login' to create a token")
}

func readTerraformCredentials() (*Credentials, error) {
	userConfigDir := os.UserHomeDir
	if runtime.GOOS == "windows" {
		userConfigDir = os.UserConfigDir
	}

	var err error
	configDir, err := userConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get the home directory: %v", err)
	}

	credentialsPath := filepath.Join(configDir, ".terraform.d", "credentials.tfrc.json")
	if debugMode {
		fmt.Println("reading credentials from:", credentialsPath)
	}

	if _, err := os.Stat(credentialsPath); errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("error checking file existence: %v", err)
	}

	fileContents, err := os.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials file: %v", err)
	}

	var creds Credentials
	err = json.Unmarshal(fileContents, &creds)
	if err != nil {
		return nil, fmt.Errorf("unable to parse JSON: %v", err)
	}

	return &creds, nil
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

func stringMapToSlice(m map[string]string) ([]string, []string) {
	keys := make([]string, len(m))
	values := make([]string, len(m))
	i := 0
	for k, v := range m {
		keys[i] = k
		values[i] = v
		i++
	}
	return keys, values
}

func workspaceListToString(wsNames []string) string {
	if len(wsNames) == 0 {
		return ""
	}

	s := ""
	if len(wsNames) > 1 {
		s = "workspaces: " + strings.Join(wsNames, ", ")
	} else {
		s = "workspace '" + wsNames[0] + "'"
	}

	return s
}
