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

	"github.com/spf13/cobra"

	cloner "github.com/sil-org/tfc-ops/v4/lib"
)

var (
	copyState                   bool
	copyVariables               bool
	applyVariableSets           bool
	differentDestinationAccount bool
	newOrganization             string
	sourceWorkspace             string
	newWorkspace                string
	newVCSTokenID               string
)

// cloneCmd represents the clone command
var cloneCmd = &cobra.Command{
	Use:   "clone",
	Short: "Clone a Workspace",
	Long:  `Clone a Terraform Cloud Workspace`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if differentDestinationAccount {

			if newOrganization == "" {
				fmt.Println("Error: The 'new-organization' '-p' flag is required for a different destination account.")
				os.Exit(1)
			}
			if newVCSTokenID == "" {
				fmt.Println("Error: The 'new-vcs-token-id' '-v' flag is required for a different destination account.")
				os.Exit(1)
			}
		}

		config := cloner.CloneConfig{
			Organization:                organization,
			NewOrganization:             newOrganization,
			SourceWorkspace:             sourceWorkspace,
			NewWorkspace:                newWorkspace,
			NewVCSTokenID:               newVCSTokenID,
			CopyState:                   copyState,
			CopyVariables:               copyVariables,
			ApplyVariableSets:           applyVariableSets,
			DifferentDestinationAccount: differentDestinationAccount,
		}

		runClone(config)
	},
}

func init() {
	workspaceCmd.AddCommand(cloneCmd)
	cloneCmd.Flags().StringVarP(
		&newOrganization,
		"new-organization",
		"p",
		"",
		`Name of the Destination Organization in Terraform Cloud`,
	)
	cloneCmd.Flags().StringVarP(
		&sourceWorkspace,
		"source-workspace",
		"s",
		"",
		requiredPrefix+`Name of the Source Workspace in Terraform Cloud`,
	)
	cloneCmd.Flags().StringVarP(
		&newWorkspace,
		"new-workspace",
		"n",
		"",
		requiredPrefix+`Name of the new Workspace in Terraform Cloud`,
	)
	cloneCmd.Flags().StringVarP(
		&newVCSTokenID,
		"new-vcs-token-id",
		"v",
		"",
		`The new organization's VCS repo's oauth-token-id`,
	)
	cloneCmd.Flags().BoolVarP(
		&copyState,
		"copyState",
		"t",
		false,
		`optional (e.g. "-t=true") whether to copy the state of the Source Workspace (only possible if copying to a new account).`,
	)
	cloneCmd.Flags().BoolVarP(
		&copyVariables,
		"copyVariables",
		"c",
		false,
		`optional (e.g. "-c=true") whether to copy the values of the Source Workspace variables.`,
	)
	cloneCmd.Flags().BoolVar(
		&applyVariableSets,
		"applyVariableSets",
		false,
		`optional, whether to apply the same variable sets to the new workspace (only for same-account clone).`,
	)
	cloneCmd.Flags().BoolVarP(
		&differentDestinationAccount,
		"differentDestinationAccount",
		"d",
		false,
		`optional (e.g. "-d=true") whether to clone to a different TF account.`,
	)
	if err := cloneCmd.MarkFlagRequired("source-workspace"); err != nil {
		errLog.Fatalln(err)
	}
	if err := cloneCmd.MarkFlagRequired("new-workspace"); err != nil {
		errLog.Fatalln(err)
	}
}

func runClone(cfg cloner.CloneConfig) {
	if readOnlyMode {
		fmt.Println("read-only mode enabled, no workspace will be created")
	}

	cfg.AtlasTokenDestination = os.Getenv("ATLAS_TOKEN_DESTINATION")
	if cfg.AtlasTokenDestination == "" {
		cfg.AtlasTokenDestination = cloner.GetToken()
		fmt.Print("Info: ATLAS_TOKEN_DESTINATION is not set, using primary credential for destination account.\n\n")
	}

	fmt.Printf("clone called using %s, %s, %s, copyState: %t, copyVariables: %t, "+
		"applyVariableSets: %t, differentDestinationAccount: %t\n",
		cfg.Organization, cfg.SourceWorkspace, cfg.NewWorkspace, cfg.CopyState, cfg.CopyVariables,
		cfg.ApplyVariableSets, cfg.DifferentDestinationAccount)

	sensitiveVars, err := cloner.CloneWorkspace(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	println("\n  **** Completed Cloning ****")
	if len(sensitiveVars) > 0 {
		fmt.Printf("Sensitive variables for %s:%s\n", cfg.Organization, cfg.NewWorkspace)
		for _, nextVar := range sensitiveVars {
			println(nextVar)
		}
	}
}
