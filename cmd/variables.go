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
	"github.com/spf13/cobra"
)

// variablesCmd represents the top level command for variables
var variablesCmd = &cobra.Command{
	Use:   "variables",
	Short: "Update or List variables",
	Long:  `Top level command to update or lists variables in all workspaces`,
	Args:  cobra.MinimumNArgs(1),
}

func init() {
	rootCmd.AddCommand(variablesCmd)
	addGlobalFlags(variablesCmd)
	variablesCmd.PersistentFlags().StringVarP(&workspace, "workspace", "w", "",
		`Name of the Workspace in Terraform Cloud`,
	)
}
