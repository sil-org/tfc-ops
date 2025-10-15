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
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	api "github.com/sil-org/tfc-ops/v5/lib"
)

var (
	keyContains   string
	valueContains string
	tabularCSV    bool
)

var variablesListCmd = &cobra.Command{
	Use:   "list",
	Short: "Report on variables",
	Long:  `Show the values of variables with a key or value containing a certain string`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if len(keyContains) == 0 && len(valueContains) == 0 {
			fmt.Println("Error: Either the 'key_contains' flag or 'value_contains flag must be set")
			fmt.Println("")
			os.Exit(1)
		}

		if tabularCSV {
			fmt.Println("workspace,key,value")
		} else {
			keyMsg := ""
			valMsg := ""

			if keyContains != "" {
				keyMsg = " key containing " + keyContains
			}

			if valueContains != "" {
				valMsg = " value containing " + valueContains
				if keyContains != "" {
					valMsg = " or value containing " + valueContains
				}
			}

			wsMsg := workspace
			if wsMsg == "" {
				wsMsg = "all workspaces"
			}
			if !tabularCSV {
				fmt.Printf("Getting variables from %s with%s%s\n", wsMsg, keyMsg, valMsg)
			}
		}
		runVariablesList()
	},
}

func init() {
	variablesCmd.AddCommand(variablesListCmd)
	variablesListCmd.Flags().StringVarP(&keyContains, "key_contains", "k", "",
		"required if value_contains is blank - string contained in the Terraform variable keys to report on")
	variablesListCmd.Flags().StringVarP(&valueContains, "value_contains", "v", "",
		"required if key_contains is blank - string contained in the Terraform variable values to report on")
	variablesListCmd.Flags().BoolVar(&tabularCSV, "csv", false,
		"output variable list in CSV format")
}

func runVariablesList() {
	if workspace != "" {
		vars, err := api.SearchVariables(organization, workspace, keyContains, valueContains)
		if err != nil {
			println(err.Error())
			return
		}
		printWorkspaceVars(workspace, vars)
		return
	}
	allData, err := api.GetAllWorkspaces(organization)
	if err != nil {
		println(err.Error())
		return
	}

	wsVars, err := api.SearchVarsInAllWorkspaces(allData, organization, keyContains, valueContains)
	if err != nil {
		println(err.Error())
		return
	}

	for ws, vs := range wsVars {
		printWorkspaceVars(ws, vs)
	}
	println()
	return
}

func printWorkspaceVars(ws string, vs []api.Var) {
	if len(vs) == 0 {
		return
	}
	if tabularCSV {
		printWorkspaceVarsCSV(ws, vs)
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.Debug)
	println()
	fmt.Printf("Workspace: %s has %v matching variable(s)\n", ws, len(vs))
	fmt.Fprintln(w, "Key \t Value \t Sensitive")
	for _, v := range vs {
		sens := ""
		if v.Sensitive {
			sens = "(sensitive)"
		}
		fmt.Fprintf(w, "%s \t %s \t %s\n", v.Key, v.Value, sens)
	}
	println()
	w.Flush()
}

func printWorkspaceVarsCSV(ws string, vs []api.Var) {
	for _, v := range vs {
		val := v.Value
		if v.Sensitive {
			val = "(sensitive)"
		}
		fmt.Printf(`"%s","%s","%s"`+"\n", escapeString(ws), escapeString(v.Key), escapeString(val))
	}
}

// escapeString escapes characters for CSV encoding, adding a backslash before a double-quote, and converting
// a newline to `\n`
func escapeString(s string) string {
	tmp := strings.Replace(s, `"`, `\"`, -1)
	return strings.Replace(tmp, "\n", `\n`, -1)
}
