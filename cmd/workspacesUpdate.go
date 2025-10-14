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
	"log"

	"github.com/spf13/cobra"

	"github.com/sil-org/tfc-ops/v4/lib"
)

const (
	flagAttribute       = "attribute"
	flagValue           = "value"
	flagWorkspaceFilter = "workspace"
)

var (
	attribute       string
	value           string
	workspaceFilter string
)

// workspaceUpdateCmd represents the workspace update command
var workspaceUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Workspaces",
	Long:  `Updates an attribute of Terraform workspaces`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating workspaces ...")
		runWorkspaceUpdate()
	},
}

func init() {
	workspaceCmd.AddCommand(workspaceUpdateCmd)

	workspaceUpdateCmd.Flags().StringVarP(&attribute, flagAttribute, "a", "",
		requiredPrefix+"Workspace attribute to update, use Terraform Cloud API workspace attribute names")
	workspaceUpdateCmd.Flags().StringVarP(&value, flagValue, "v", "",
		requiredPrefix+"Value")
	workspaceUpdateCmd.Flags().StringVarP(&workspaceFilter, flagWorkspaceFilter, "w", "",
		requiredPrefix+"Partial workspace name to search across all workspaces")

	requiredFlags := []string{flagAttribute, flagValue, flagWorkspaceFilter}
	for _, flag := range requiredFlags {
		if err := workspaceUpdateCmd.MarkFlagRequired(flag); err != nil {
			panic("MarkFlagRequired failed with error: " + err.Error())
		}
	}
}

func runWorkspaceUpdate() {
	if err := lib.UpdateWorkspace(lib.WorkspaceUpdateParams{
		Organization:    organization,
		WorkspaceFilter: workspaceFilter,
		Attribute:       attribute,
		Value:           value,
	}); err != nil {
		log.Fatalln(err.Error())
	}
}
