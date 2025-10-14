package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/sil-org/tfc-ops/v4/lib"
)

var (
	workspace             string
	variableSearchString  string
	newVariableValue      string
	searchOnVariableValue bool
	addKeyIfNotFound      bool
	sensitiveVariable     bool
)

// cloneCmd represents the clone command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update/add a variable in a Workspace",
	Long:  `Update or add a variable in a Terraform Cloud Workspace based on a complete case-insensitive match`,
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if addKeyIfNotFound && searchOnVariableValue {
			fmt.Println("Error: The 'add-key-if-not-found' flag may not be used with the 'search-on-variable-value' flag")
			os.Exit(1)
		}
		config := lib.UpdateConfig{
			Organization:          organization,
			Workspace:             workspace,
			SearchString:          variableSearchString,
			NewValue:              newVariableValue,
			SearchOnVariableValue: searchOnVariableValue,
			AddKeyIfNotFound:      addKeyIfNotFound,
			SensitiveVariable:     sensitiveVariable,
		}
		if workspace == "" {
			runVariablesUpdateAll(config)
		} else {
			runVariablesUpdate(config)
		}
	},
}

func init() {
	variablesCmd.AddCommand(updateCmd)
	updateCmd.Flags().StringVarP(
		&variableSearchString,
		"variable-search-string",
		"s",
		"",
		requiredPrefix+`The string to match in the current variables (either in the Key or Value - see other flags)`,
	)
	updateCmd.Flags().StringVarP(
		&newVariableValue,
		"new-variable-value",
		"n",
		"",
		requiredPrefix+`The desired new value of the variable`,
	)
	updateCmd.Flags().BoolVarP(
		&addKeyIfNotFound,
		"add-key-if-not-found",
		"a",
		false,
		`optional (e.g. "-a=true") whether to add a new variable if a matching key is not found.`,
	)
	updateCmd.Flags().BoolVarP(
		&searchOnVariableValue,
		"search-on-variable-value",
		"v",
		false,
		`optional (e.g. "-v=true") whether to do the search based on the value of the variables. (Must be false if add-key-if-not-found is true`,
	)
	updateCmd.Flags().BoolVarP(
		&sensitiveVariable,
		"sensitive-variable",
		"x",
		false,
		`optional (e.g. "-x=true") make the variable sensitive.`,
	)
	if err := updateCmd.MarkFlagRequired("variable-search-string"); err != nil {
		errLog.Fatalln(err)
	}
	if err := updateCmd.MarkFlagRequired("new-variable-value"); err != nil {
		errLog.Fatalln(err)
	}
}

func runVariablesUpdate(cfg lib.UpdateConfig) {
	if cfg.AddKeyIfNotFound {
		if cfg.SearchOnVariableValue {
			println("update variable aborted. Because addKeyIfNotFound was true, searchOnVariableValue must be set to false")
			return
		}
		cfg.SearchOnVariableValue = false
	}

	if readOnlyMode {
		println("\n ****  READ ONLY MODE  ****")
	}
	fmt.Printf("update variable called using %s, %s, search string: %s, new value: %s, add-key-if-not-found: %t, search-on-variable-value: %t\n",
		cfg.Organization, cfg.Workspace, cfg.SearchString, cfg.NewValue, cfg.AddKeyIfNotFound, cfg.SearchOnVariableValue)

	message, err := lib.AddOrUpdateVariable(cfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	println("\n  **** Completed Updating ****")
	println(message)
}

func runVariablesUpdateAll(cfg lib.UpdateConfig) {
	allData, err := lib.GetAllWorkspaces(organization)
	for _, ws := range allData {
		value, err := ws.AttributeByLabel(strings.Trim("name", " "))
		fmt.Printf("Do you want to update the variable %s across the workspace: %s\n\n", variableSearchString, value)
		if awaitUserResponse() {
			cfg.Workspace = value
			runVariablesUpdate(cfg)

			if err != nil {
				fmt.Println("\n", err.Error())
				return
			}
		}
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

func awaitUserResponse() bool {
	prompt := promptui.Select{
		Label: "Select[Yes/No]",
		Items: []string{"No", "Yes"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		errLog.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}
