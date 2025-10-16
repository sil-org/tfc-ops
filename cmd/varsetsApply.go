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

var varsetsApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply Variable Set to Workspaces",
	Long:  `Apply an existing variable set to workspaces`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		runVarsetsApply(variableSet)
	},
}

func init() {
	varsetsCmd.AddCommand(varsetsApplyCmd)

	varsetsApplyCmd.Flags().StringVarP(&variableSet, "set", "s", "",
		requiredPrefix+"Terraform variable set to add")
	if err := varsetsApplyCmd.MarkFlagRequired("set"); err != nil {
		errLog.Fatalln("failed to mark 'set' as a required flag on varsetsApplyCmd")
	}

	varsetsApplyCmd.Flags().StringVarP(&workspace, "workspace", "w", "",
		"Name of the Workspace in Terraform Cloud")

	varsetsApplyCmd.Flags().StringVar(&workspaceFilter, "workspace-filter", "",
		"Partial workspace name to search across all workspaces")
}

func runVarsetsApply(name string) {
	if readOnlyMode {
		fmt.Println("Read only mode enabled. No variable set will be applied.")
	}

	if workspace == "" && workspaceFilter == "" {
		errLog.Fatalln("Either --workspace or --workspace-filter must be specified.")
	}

	var workspaceNames map[string]string
	if workspace != "" {
		w, err := lib.GetWorkspaceByName(organization, workspace)
		if err != nil {
			errLog.Fatalf("error getting workspace from Terraform: %s", err)
		}
		workspaceNames = map[string]string{w.ID: workspace}
	} else {
		workspaceNames = lib.FindWorkspaces(organization, workspaceFilter)
		if len(workspaceNames) == 0 {
			errLog.Fatalf("no workspaces match the filter '%s'", workspaceFilter)
		}
	}

	_ = applyVariableSet(organization, name, workspaceNames)
	return
}

func applyVariableSet(org, vsName string, workspaceNames map[string]string) bool {
	vs, err := lib.GetVariableSet(org, vsName)
	if err != nil {
		errLog.Fatalf("Error retrieving variable set: %s", err)
	}
	if vs == nil {
		errLog.Fatalf("No variable set matches the name given (%s)", vsName)
	}

	wsIDs, wsNames := stringMapToSlice(workspaceNames)

	fmt.Printf("Applying variable set '%s' to %s\n", vs.Attributes.Name, workspaceListToString(wsNames))
	if err = lib.ApplyVariableSet(vs.ID, wsIDs); err != nil {
		errLog.Fatalf("Error while applying variable set: %s", err)
	}
	return true
}
