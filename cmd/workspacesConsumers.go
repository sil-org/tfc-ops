// Copyright © 2023 SIL Global
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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sil-org/tfc-ops/v5/lib"
)

const (
	flagConsumers = "consumers"
	flagWorkspace = "workspace"
)

func addConsumersCommand(parentCommand *cobra.Command) {
	var consumers string
	workspaceConsumersCmd := &cobra.Command{
		Use:   flagConsumers,
		Short: "Manage workspace remote state consumers",
		Long:  `Add to workspace remote state consumers. (Possible future capability: list, replace, delete)`,
		Args:  cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			runWorkspaceConsumers(consumers)
		},
	}

	parentCommand.AddCommand(workspaceConsumersCmd)

	workspaceConsumersCmd.Flags().StringVarP(&workspace, flagWorkspace, "w", "",
		requiredPrefix+"Partial workspace name to search across all workspaces")

	workspaceConsumersCmd.Flags().StringVar(&consumers, flagConsumers, "",
		requiredPrefix+"List of remote state consumer workspaces, comma-separated")

	requiredFlags := []string{flagConsumers, flagWorkspace}
	for _, flag := range requiredFlags {
		if err := workspaceConsumersCmd.MarkFlagRequired(flag); err != nil {
			panic("MarkFlagRequired failed with error: " + err.Error())
		}
	}
}

func runWorkspaceConsumers(consumers string) {
	workspaceData, err := lib.GetWorkspaceData(organization, workspace)
	if err != nil {
		log.Fatalln("workspace consumers", err)
	}

	consumersList := strings.Split(consumers, ",")
	consumerIDs := make([]string, len(consumersList))
	for i, consumer := range consumersList {
		consumerData, err := lib.GetWorkspaceData(organization, consumer)
		if err != nil {
			log.Fatalln("workspace consumers", err)
		}
		consumerIDs[i] = consumerData.Data.ID
	}

	fmt.Printf("Adding to %s: %s", workspace, consumers)
	if !readOnlyMode {
		if err := lib.AddRemoteStateConsumers(workspaceData.Data.ID, consumerIDs); err != nil {
			log.Fatalln("workspace consumers", err)
		}
	}
}
