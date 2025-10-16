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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sil-org/tfc-ops/v5/lib"
)

var attributes string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List Workspaces",
	Long:  `Lists the TF workspaces with (some of) their attributes`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Getting list of workspaces ...")
		runList()
	},
}

func init() {
	workspaceCmd.AddCommand(listCmd)
	const flagAttributes = "attributes"
	listCmd.Flags().StringVarP(&attributes, flagAttributes, "a", "",
		requiredPrefix+"Workspace attributes to list, use Terraform Cloud API workspace attribute names")
	_ = listCmd.MarkFlagRequired(flagAttributes)
}

func runList() {
	allAttrs := strings.Split(attributes, ",")
	allData, err := lib.GetWorkspaceAttributes(organization, allAttrs)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(strings.Join(allAttrs, ", "))
	for _, ws := range allData {
		fmt.Println(strings.Join(ws, ", "))
	}
}
