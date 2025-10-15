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
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sil-org/tfc-ops/v5/lib"
)

var variablesAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add new variable (if not already present)",
	Long:  `Add variable in matching workspace. Will not update existing variable.`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		runVariablesAdd()
	},
}

func init() {
	variablesCmd.AddCommand(variablesAddCmd)
	variablesAddCmd.Flags().StringVarP(&key, "key", "k", "",
		requiredPrefix+"Terraform variable key")
	if err := variablesAddCmd.MarkFlagRequired("key"); err != nil {
		errLog.Fatalln("failed to mark 'key' as a required flag on variablesAddCmd")
	}
	variablesAddCmd.Flags().StringVarP(&value, "value", "v", "",
		requiredPrefix+"Terraform variable value")
	if err := variablesAddCmd.MarkFlagRequired("value"); err != nil {
		errLog.Fatalln("failed to mark 'value' as a required flag on variablesAddCmd")
	}
}

func runVariablesAdd() {
	if readOnlyMode {
		fmt.Println("Read only mode enabled. No variables will be added.")
	}

	if workspace != "" {
		addWorkspaceVar(organization, workspace, key, value)
		return
	}

	fmt.Printf("Adding variables with key '%s' and value '%s' to all workspaces...\n", key, value)
	allWorkspaces, err := lib.GetAllWorkspaces(organization)
	if err != nil {
		println(err.Error())
		return
	}

	for _, w := range allWorkspaces {
		addWorkspaceVar(organization, w.Attributes.Name, key, value)
	}
	return
}

func addWorkspaceVar(org, ws, key, value string) {
	if v, err := lib.GetWorkspaceVar(org, ws, key); err != nil {
		errLog.Fatalf("failure checking for existence of variable '%s' in workspace '%s', %s\n", key, ws, err)
	} else if v != nil {
		errLog.Fatalf("'%s' already exists in '%s'. Use 'variable update' command to change the value.\n", key, ws)
	}

	fmt.Printf("Workspace %s: Adding variable %s = %s\n", ws, key, value)
	if !readOnlyMode {
		if _, err := lib.AddOrUpdateVariable(lib.UpdateConfig{
			Organization:          organization,
			Workspace:             ws,
			SearchString:          key,
			NewValue:              value,
			AddKeyIfNotFound:      true,
			SearchOnVariableValue: false,
			SensitiveVariable:     false,
		}); err != nil {
			errLog.Fatalf("failed to add variable '%s' in workspace '%s', %s\n", key, ws, err)
		}
	}
	return
}
