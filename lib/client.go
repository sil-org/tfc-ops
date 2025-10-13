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

package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type UpdateConfig struct {
	Organization          string
	Workspace             string
	SearchString          string //  must be an exact case-insensitive match (i.e. not a partial match)
	NewValue              string
	AddKeyIfNotFound      bool // If true, then SearchOnVariableValue will be treated as false
	SearchOnVariableValue bool // If false, then will filter on variable key
	SensitiveVariable     bool // Whether to mark the variable as sensitive
}

type CloneConfig struct {
	Organization                string
	NewOrganization             string
	SourceWorkspace             string
	NewWorkspace                string
	NewVCSTokenID               string
	AtlasTokenDestination       string
	CopyState                   bool
	CopyVariables               bool
	ApplyVariableSets           bool
	DifferentDestinationAccount bool
}

// Var is what is returned by the api for one variable
type Var struct {
	ID        string `json:"-"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	Sensitive bool   `json:"sensitive"`
	Category  string `json:"category"`
	Hcl       bool   `json:"hcl"`
}

// VarsResponse is what is returned by the api when requesting the variables of a workspace
type VarsResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Type          string `json:"type"`
		Variable      Var    `json:"attributes"`
		Relationships struct {
			Configurable struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"configurable"`
		} `json:"relationships"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
}

// Workspace is what is returned by the api for each workspace
type Workspace struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Name             string    `json:"name"`
		Environment      string    `json:"environment"`
		AutoApply        bool      `json:"auto-apply"`
		Locked           bool      `json:"locked"`
		CreatedAt        time.Time `json:"created-at"`
		WorkingDirectory string    `json:"working-directory"`
		VCSRepo          struct {
			Branch            string `json:"branch"`
			Identifier        string `json:"identifier"`
			DisplayIdentifier string `json:"display-identifier"`
			TokenID           string `json:"oauth-token-id"`
		} `json:"vcs-repo"`
		StructuredRunOutputEnabled bool   `json:"structured-run-output-enabled"`
		TerraformVersion           string `json:"terraform-version"`
		Permissions                struct {
			CanUpdate         bool `json:"can-update"`
			CanDestroy        bool `json:"can-destroy"`
			CanQueueDestroy   bool `json:"can-queue-destroy"`
			CanQueueRun       bool `json:"can-queue-run"`
			CanUpdateVariable bool `json:"can-update-variable"`
			CanLock           bool `json:"can-lock"`
			CanReadSettings   bool `json:"can-read-settings"`
		} `json:"permissions"`
		Actions struct {
			IsDestroyable bool `json:"is-destroyable"`
		} `json:"actions"`
	} `json:"attributes"`
	Relationships struct {
		Organization struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"organization"`
		LatestRun struct {
			Data any `json:"data"`
		} `json:"latest-run"`
		CurrentRun struct {
			Data any `json:"data"`
		} `json:"current-run"`
	} `json:"relationships"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

const (
	WsAttrID                   = "id"
	WsAttrAutoApply            = "auto-apply"
	WsAttrCreatedAt            = "created-at"
	WsAttrEnvironment          = "environment"
	WsAttrName                 = "name"
	WsAttrStructuredRunOutput  = "structured-run-output-enabled"
	WsAttrTerraformVersion     = "terraform-version"
	WsAttrVcsDisplayIdentifier = "vcs-repo.display-identifier"
	WsAttrVcsTokenID           = "vcs-repo.oauth-token-id"
	WsAttrWorkingDirectory     = "working-directory"
)

func (v *Workspace) AttributeByLabel(label string) (string, error) {
	switch strings.ToLower(label) {
	case WsAttrID:
		return v.ID, nil
	case WsAttrAutoApply:
		return fmt.Sprintf("%v", v.Attributes.AutoApply), nil
	case WsAttrCreatedAt, "createdat":
		return v.Attributes.CreatedAt.String(), nil
	case WsAttrEnvironment:
		return v.Attributes.Environment, nil
	case WsAttrName:
		return v.Attributes.Name, nil
	case WsAttrStructuredRunOutput:
		return fmt.Sprintf("%v", v.Attributes.StructuredRunOutputEnabled), nil
	case WsAttrTerraformVersion, "terraformversion":
		return v.Attributes.TerraformVersion, nil
	case "vcsrepo":
		return v.Attributes.VCSRepo.Identifier, nil
	case WsAttrVcsDisplayIdentifier:
		return v.Attributes.VCSRepo.DisplayIdentifier, nil
	case WsAttrVcsTokenID:
		return v.Attributes.VCSRepo.TokenID, nil
	case WsAttrWorkingDirectory, "workingdirectory":
		return v.Attributes.WorkingDirectory, nil
	}

	return "", fmt.Errorf("Attribute label not valid: %s", label)
}

// WorkspaceJSON is what is returned by the api when requesting the data for a workspace
type WorkspaceJSON struct {
	Data Workspace `json:"data"`
}

// WorkspaceList is returned by the API when requesting a list of workspaces
type WorkspaceList struct {
	Data []Workspace `json:"data"`
}

// TeamWorkspaceData is what is returned by the api for one team access object for a workspace
type TeamWorkspaceData struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Access string `json:"access"`
	} `json:"attributes"`
	Relationships struct {
		Team struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
			Links struct {
				Related string `json:"related"`
			} `json:"links"`
		} `json:"team"`
		Workspace struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
			Links struct {
				Related string `json:"related"`
			} `json:"links"`
		} `json:"workspace"`
	} `json:"relationships"`
	Links struct {
		Self string `json:"self"`
	} `json:"links"`
}

// AllTeamWorkspaceData is what is returned by the api when requesting the team access data for a workspace
type AllTeamWorkspaceData struct {
	Data []TeamWorkspaceData `json:"data"`
}

type WorkspaceUpdateParams struct {
	Organization    string
	WorkspaceFilter string
	Attribute       string
	Value           string
}

// ConvertHCLVariable changes a Var struct in place by escaping
// the double quotes and line endings in the Value attribute
func ConvertHCLVariable(tfVar *Var) {
	if !tfVar.Hcl {
		return
	}

	tfVar.Value = strings.Replace(tfVar.Value, `"`, `\"`, -1)
	tfVar.Value = strings.Replace(tfVar.Value, "\n", "\\n", -1)
}

// GetCreateVariablePayload returns the json needed to make a Post to the
// Terraform vars api
func GetCreateVariablePayload(organization, workspaceName string, tfVar Var) string {
	return fmt.Sprintf(`
{
  "data": {
    "type":"vars",
    "attributes": {
      "key":"%s",
      "value":"%s",
      "category":"terraform",
      "hcl":%t,
      "sensitive":%t
    }
  },
  "filter": {
    "organization": {
      "name":"%s"
    },
    "workspace": {
      "name":"%s"
    }
  }
}
`, tfVar.Key, tfVar.Value, tfVar.Hcl, tfVar.Sensitive, organization, workspaceName)
}

// GetUpdateVariablePayload returns the json needed to make a Post to the
// Terraform vars api
func GetUpdateVariablePayload(organization, workspaceName, variableID string, tfVar Var) string {
	return fmt.Sprintf(`
{
  "data": {
    "id":"%s",
    "type":"vars",
    "attributes": {
      "key":"%s",
      "value":"%s",
      "category":"terraform",
      "description":"",
      "hcl":%t,
      "sensitive":%t
    }
  },
  "filter": {
    "organization": {
      "name":"%s"
    },
    "workspace": {
      "name":"%s"
    }
  }
}
`, variableID, tfVar.Key, tfVar.Value, tfVar.Hcl, tfVar.Sensitive, organization, workspaceName)
}

// OrganizationExists returns whether an organization with the given name exists
func OrganizationExists(organization string) (bool, error) {
	if organization == "" {
		return false, fmt.Errorf("OrganizationExists: organization is required")
	}
	u := NewTfcUrl("/organizations/" + organization)

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	// Status 200 indicates the organization exists
	return resp.StatusCode == http.StatusOK, nil
}

// GetAllWorkspaces retrieves all workspaces from Terraform Cloud and returns a list of Workspace objects
func GetAllWorkspaces(organization string) ([]Workspace, error) {
	u := NewTfcUrl(fmt.Sprintf("/organizations/%s/workspaces", organization))
	u.SetParam(paramPageSize, strconv.Itoa(pageSize))

	allWsData := []Workspace{}

	for page := 1; ; page++ {
		u.SetParam(paramPageNumber, strconv.Itoa(page))
		nextWsData, err := getWorkspacePage(u.String())
		if err != nil {
			return nil, fmt.Errorf("error getting workspace data for %s: %s", organization, err)
		}
		allWsData = append(allWsData, nextWsData.Data...)

		// If there isn't a whole page of contents, then we're on the last one.
		if len(nextWsData.Data) < pageSize {
			break
		}
	}
	return allWsData, nil
}

func getWorkspacePage(url string) (WorkspaceList, error) {
	resp, err := callAPI(http.MethodGet, url, "", nil)
	if err != nil {
		return WorkspaceList{}, err
	}
	defer resp.Body.Close()

	var nextWsData WorkspaceList
	if err := json.NewDecoder(resp.Body).Decode(&nextWsData); err != nil {
		return WorkspaceList{}, fmt.Errorf("json decode error: %s", err)
	}
	return nextWsData, nil
}

func GetWorkspaceData(organization, workspaceName string) (WorkspaceJSON, error) {
	if organization == "" {
		return WorkspaceJSON{}, fmt.Errorf("GetWorkspaceData: organization is required")
	}
	if workspaceName == "" {
		return WorkspaceJSON{}, fmt.Errorf("GetWorkspaceData: workspace is required")
	}
	u := NewTfcUrl(fmt.Sprintf(
		"/organizations/%s/workspaces/%s",
		organization,
		workspaceName,
	))

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return WorkspaceJSON{}, err
	}
	defer resp.Body.Close()

	var wsData WorkspaceJSON
	if err := json.NewDecoder(resp.Body).Decode(&wsData); err != nil {
		return WorkspaceJSON{}, fmt.Errorf("Error getting workspace data for %s:%s\n%s", organization, workspaceName, err.Error())
	}

	return wsData, nil
}

// GetWorkspaceVar retrieves the variables from a Workspace and returns the Var that matches the given key
func GetWorkspaceVar(organization, wsName, key string) (*Var, error) {
	vars, err := GetVarsFromWorkspace(organization, wsName)
	if err != nil {
		return nil, fmt.Errorf("Error getting variables for %s:%s\n%w", organization, wsName, err)
	}

	for _, v := range vars {
		if v.Key == key {
			found := v
			return &found, nil
		}
	}
	return nil, nil
}

// GetVarsFromWorkspace returns a list of Terraform variables for a given workspace
func GetVarsFromWorkspace(organization, workspaceName string) ([]Var, error) {
	orgExists, err := OrganizationExists(organization)
	if err != nil {
		return nil, fmt.Errorf("GetVarsFromWorkspace: organization is required: %w", err)
	} else if !orgExists {
		return nil, fmt.Errorf("GetVarsFromWorkspace: invalid organization %s", organization)
	}

	if workspaceName == "" {
		return nil, fmt.Errorf("GetVarsFromWorkspace: workspace is required")
	}

	u := NewTfcUrl("/vars")
	u.SetParam(paramFilterOrganizationName, organization)
	u.SetParam(paramFilterWorkspaceName, workspaceName)

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var varsResp VarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&varsResp); err != nil {
		return nil, fmt.Errorf("Error getting variables for %s:%s ...\n%s", organization, workspaceName, err.Error())
	}

	variables := []Var{}
	for _, data := range varsResp.Data {
		data.Variable.ID = data.ID // push the ID down into the Variable for future reference
		variables = append(variables, data.Variable)
	}

	return variables, nil
}

// DeleteVariable deletes a variable from a workspace
func DeleteVariable(variableID string) {
	u := NewTfcUrl("/vars/" + variableID)

	resp, err := callAPI(http.MethodDelete, u.String(), "", nil)
	if err != nil {
		log.Fatalln(err)
	}
	_ = resp.Body.Close()
}

// SearchVarsInAllWorkspaces returns all the variables that match the search terms 'keyContains' and 'valueContains'
// in all workspaces given. The return value is a map of variable lists with the workspace name as the key.
func SearchVarsInAllWorkspaces(wsData []Workspace, organization, keyContains, valueContains string) (map[string][]Var, error) {
	allVars := map[string][]Var{}

	for _, ws := range wsData {
		wsName := ws.Attributes.Name

		wsVars, err := SearchVariables(organization, wsName, keyContains, valueContains)
		if err != nil {
			return nil, err
		}
		allVars[wsName] = wsVars
	}

	return allVars, nil
}

// SearchVariables returns a list of variables in the given workspace that match the search terms
// 'keyContains' and 'valueContains'
func SearchVariables(organization, wsName, keyContains, valueContains string) ([]Var, error) {
	vars, err := GetVarsFromWorkspace(organization, wsName)
	if err != nil {
		err := fmt.Errorf("Error getting variables for %s:%s\n%s", organization, wsName, err.Error())
		return nil, err
	}

	var wsVars []Var

	for _, v := range vars {
		if keyContains != "" && strings.Contains(v.Key, keyContains) {
			wsVars = append(wsVars, v)
			continue
		}
		if valueContains != "" && strings.Contains(v.Value, valueContains) {
			wsVars = append(wsVars, v)
		}
	}

	return wsVars, nil
}

// GetTeamAccessFrom returns the team access data from an existing workspace
func GetTeamAccessFrom(workspaceID string) (AllTeamWorkspaceData, error) {
	u := NewTfcUrl("/team-workspaces")
	u.SetParam(paramFilterWorkspaceID, workspaceID)

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return AllTeamWorkspaceData{}, err
	}
	defer resp.Body.Close()

	var allTeamData AllTeamWorkspaceData
	if err := json.NewDecoder(resp.Body).Decode(&allTeamData); err != nil {
		return AllTeamWorkspaceData{}, fmt.Errorf("Error getting team workspace data for %s\n%s", workspaceID, err.Error())
	}

	return allTeamData, nil
}

func getAssignTeamAccessPayload(accessLevel, workspaceID, teamID string) string {
	return fmt.Sprintf(`
{
  "data": {
    "attributes": {
      "access":"%s"
    },
    "relationships": {
      "workspace": {
        "data": {
          "type":"workspaces",
          "id":"%s"
        }
      },
      "team": {
        "data": {
          "type":"teams",
          "id":"%s"
        }
      }
    },
    "type":"team-workspaces"
  }
}
`, accessLevel, workspaceID, teamID)
}

// AssignTeamAccess assigns the requested team access to a workspace on Terraform Cloud
func AssignTeamAccess(workspaceID string, allTeamData AllTeamWorkspaceData) {
	url := fmt.Sprintf(baseURL + "/team-workspaces")

	for _, teamData := range allTeamData.Data {
		postData := getAssignTeamAccessPayload(
			teamData.Attributes.Access,
			workspaceID,
			teamData.Relationships.Team.Data.ID,
		)

		resp, err := callAPI(http.MethodPost, url, postData, nil)
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()
	}
}

// CreateVariable makes a Terraform vars API POST to create a variable
// for a given organization and workspace
func CreateVariable(organization, workspaceName string, tfVar Var) {
	url := baseURL + "/vars"

	ConvertHCLVariable(&tfVar)

	postData := GetCreateVariablePayload(organization, workspaceName, tfVar)

	resp, err := callAPI(http.MethodPost, url, postData, nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
}

// CreateAllVariables makes several Terraform vars API POSTs to create
// variables for a given organization and workspace
func CreateAllVariables(organization, workspaceName string, tfVars []Var) {
	for _, nextVar := range tfVars {
		CreateVariable(organization, workspaceName, nextVar)
	}
}

// GetCreateWorkspacePayload returns the JSON needed to make a POST to the
// Terraform workspaces API
func GetCreateWorkspacePayload(oc OpsConfig, vcsTokenID string) string {
	jsonObj := gabs.Wrap(map[string]any{
		"data": map[string]any{
			"type": "workspaces",
		},
	})
	_, _ = jsonObj.SetP(oc.NewName, "data.attributes.name")
	_, _ = jsonObj.SetP(oc.TerraformVersion, "data.attributes.terraform_version")
	_, _ = jsonObj.SetP(oc.Directory, "data.attributes.working-directory")
	if vcsTokenID != "" {
		_, _ = jsonObj.SetP(oc.RepoID, "data.attributes.vcs-repo.identifier")
		_, _ = jsonObj.SetP(vcsTokenID, "data.attributes.vcs-repo.oauth-token-id")
		_, _ = jsonObj.SetP(oc.Branch, "data.attributes.vcs-repo.branch")
		_, _ = jsonObj.SetP(true, "data.attributes.vcs-repo.default-branch")
	}

	return jsonObj.String()
}

// UpdateVariable makes a Terraform vars API call to update a variable
// for a given organization and workspace
func UpdateVariable(organization, workspaceName, variableID string, tfVar Var) {
	url := fmt.Sprintf(baseURL+"/vars/%s", variableID)

	ConvertHCLVariable(&tfVar)

	patchData := GetUpdateVariablePayload(organization, workspaceName, variableID, tfVar)

	resp, err := callAPI(http.MethodPatch, url, patchData, nil)
	if err != nil {
		log.Fatalln(err)
	}

	defer resp.Body.Close()
}

// CreateWorkspace makes a Terraform workspaces API call to create a
// workspace for a given organization, including setting up its VCS repo integration
func CreateWorkspace(oc OpsConfig, vcsTokenID string) (string, error) {
	url := fmt.Sprintf(
		baseURL+"/organizations/%s/workspaces",
		oc.NewOrg,
	)

	postData := GetCreateWorkspacePayload(oc, vcsTokenID)

	resp, err := callAPI(http.MethodPost, url, postData, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var wsData WorkspaceJSON
	if err := json.NewDecoder(resp.Body).Decode(&wsData); err != nil {
		return "", fmt.Errorf("error getting created workspace data: %s\n", err)
	}
	return wsData.Data.ID, nil
}

// CreateWorkspace2 makes a Terraform workspaces API call to create a workspace for a given organization, including
// setting up its VCS repo integration. Returns the properties of the new workspace.
func CreateWorkspace2(oc OpsConfig, vcsTokenID string) (Workspace, error) {
	url := fmt.Sprintf(
		baseURL+"/organizations/%s/workspaces",
		oc.NewOrg,
	)

	postData := GetCreateWorkspacePayload(oc, vcsTokenID)

	resp, err := callAPI(http.MethodPost, url, postData, nil)
	if err != nil {
		return Workspace{}, err
	}
	defer resp.Body.Close()

	var wsData WorkspaceJSON
	if err := json.NewDecoder(resp.Body).Decode(&wsData); err != nil {
		return Workspace{}, fmt.Errorf("error getting created workspace data: %s\n", err)
	}
	return wsData.Data, nil
}

// RunTFInit ...
//   - removes old terraform.tfstate files
//   - runs terraform init with old versions
//   - runs terraform init with new version
//
// NOTE: This procedure can be used to copy/migrate a workspace's state to a new one.
// (see the -backend-config mention below and the backend.tf file in this repo)
func RunTFInit(oc OpsConfig, tfTokenDestination string) error {
	var tfInit string
	var err error
	var osCmd *exec.Cmd
	var stderr bytes.Buffer

	stateFile := ".terraform"

	// Remove previous state file, if it exists
	_, err = os.Stat(stateFile)
	if err == nil {
		err = os.RemoveAll(stateFile)
		if err != nil {
			return err
		}
	}

	tfInit = fmt.Sprintf(`-backend-config=name=%s/%s`, oc.SourceOrg, oc.SourceName)

	osCmd = exec.Command("terraform", "init", tfInit)
	osCmd.Stderr = &stderr

	err = osCmd.Run()
	if err != nil {
		println("Error with Legacy: " + tfInit)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	SetToken(tfTokenDestination)

	// Run tf init with new version
	tfInit = fmt.Sprintf(`-backend-config=name=%s/%s`, oc.NewOrg, oc.NewName)
	osCmd = exec.Command("terraform", "init", tfInit)
	osCmd.Stderr = &stderr

	// Needed to run the command interactively, in order to allow for an automated reply
	cmdStdin, err := osCmd.StdinPipe()
	if err != nil {
		println("Error with StdinPipe: " + tfInit)
		return err
	}

	err = osCmd.Start()
	if err != nil {
		return err
	}

	defer cmdStdin.Close()
	io.Copy(cmdStdin, bytes.NewBufferString("yes\n"))

	//  Answer "yes" to the question about creating the new state
	err = osCmd.Wait()
	if err != nil {
		println("Error waiting for new tf init: " + tfInit)
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}

	return nil
}

// CloneWorkspace gets the data, variables and team access data for an existing Terraform Cloud workspace
// and then creates a clone of it with the same data.
//
// If the copyVariables param is set to true, then all the non-sensitive variable values will be added to the new
// workspace.  Otherwise, they will be set to "REPLACE_THIS_VALUE"
func CloneWorkspace(cfg CloneConfig) ([]string, error) {
	sourceWsData, err := GetWorkspaceData(cfg.Organization, cfg.SourceWorkspace)
	if err != nil {
		return nil, err
	}

	variables, err := GetVarsFromWorkspace(cfg.Organization, cfg.SourceWorkspace)
	if err != nil {
		return nil, err
	}

	if !cfg.DifferentDestinationAccount {
		cfg.NewOrganization = cfg.Organization
		cfg.NewVCSTokenID = sourceWsData.Data.Attributes.VCSRepo.Identifier
	}

	oc := OpsConfig{
		SourceOrg:        cfg.Organization,
		SourceName:       sourceWsData.Data.Attributes.Name,
		NewOrg:           cfg.NewOrganization,
		NewName:          cfg.NewWorkspace,
		TerraformVersion: sourceWsData.Data.Attributes.TerraformVersion,
		RepoID:           sourceWsData.Data.Attributes.VCSRepo.Identifier,
		Branch:           sourceWsData.Data.Attributes.VCSRepo.Branch,
		Directory:        sourceWsData.Data.Attributes.WorkingDirectory,
	}

	sensitiveVars := []string{}
	sensitiveValue := "TF_ENTERPRISE_SENSITIVE_VAR"
	defaultValue := "REPLACE_THIS_VALUE"

	tfVars := []Var{}
	var tfVar Var

	for _, nextVar := range variables {
		if cfg.CopyVariables {
			tfVar = Var{
				Key:   nextVar.Key,
				Value: nextVar.Value,
				Hcl:   nextVar.Hcl,
			}
		} else {
			tfVar = Var{
				Key:   nextVar.Key,
				Value: defaultValue,
				Hcl:   nextVar.Hcl,
			}
		}
		if nextVar.Value == sensitiveValue {
			sensitiveVars = append(sensitiveVars, nextVar.Key)
		}
		tfVars = append(tfVars, tfVar)
	}

	if config.readOnly {
		return sensitiveVars, nil
	}

	if cfg.DifferentDestinationAccount {
		// save primary token and set destination token to create the workspace and variables
		primaryToken := config.token
		SetToken(cfg.AtlasTokenDestination)
		_, err := CreateWorkspace(oc, cfg.NewVCSTokenID)
		if err != nil {
			return nil, err
		}
		CreateAllVariables(oc.NewOrg, oc.NewName, tfVars)
		SetToken(primaryToken)

		if cfg.CopyState {
			if err := RunTFInit(oc, cfg.AtlasTokenDestination); err != nil {
				return sensitiveVars, err
			}
		}

		return sensitiveVars, nil
	}

	destWsProps, err := CreateWorkspace2(oc, sourceWsData.Data.Attributes.VCSRepo.TokenID)
	if err != nil {
		return nil, fmt.Errorf("failed to create new workspace: %w", err)
	}

	err = copyVariableSetList(sourceWsData.Data.ID, destWsProps.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to clone variable sets: %w", err)
	}

	CreateAllVariables(oc.NewOrg, oc.NewName, tfVars)

	// Get Team Access Data for source Workspace
	allTeamData, err := GetTeamAccessFrom(sourceWsData.Data.ID)
	if err != nil {
		return sensitiveVars, err
	}

	// Get new Workspace data for its ID
	newWsData, err := GetWorkspaceData(cfg.Organization, cfg.NewWorkspace)
	if err != nil {
		return sensitiveVars, err
	}

	AssignTeamAccess(newWsData.Data.ID, allTeamData)

	return sensitiveVars, nil
}

// AddOrUpdateVariable adds or updates an existing Terraform Cloud workspace variable
// If the copyVariables param is set to true, then all the non-sensitive variable values will be added to the new
// workspace.  Otherwise, they will be set to "REPLACE_THIS_VALUE"
func AddOrUpdateVariable(cfg UpdateConfig) (string, error) {
	variables, err := GetVarsFromWorkspace(cfg.Organization, cfg.Workspace)
	if err != nil {
		return "", err
	}

	loweredSearchString := strings.ToLower(cfg.SearchString)

	for _, nextVar := range variables {
		oldValue := nextVar.Value
		if cfg.SearchOnVariableValue {
			if strings.ToLower(nextVar.Value) != loweredSearchString {
				continue
			}
			// Found a match
			tfVar := Var{Key: nextVar.Key, Value: cfg.NewValue, Hcl: false, Sensitive: cfg.SensitiveVariable}
			if !config.readOnly {
				UpdateVariable(cfg.Organization, cfg.Workspace, nextVar.ID, tfVar)
			}
			return fmt.Sprintf("Replaced the value of %s from %s to %s", nextVar.Key, oldValue, cfg.NewValue), nil
		}

		// Search on variable key, since search on value is not true
		if strings.ToLower(nextVar.Key) != loweredSearchString {
			continue
		}

		// Found a match
		// Only add if there isn't a match
		if cfg.AddKeyIfNotFound {
			return "", errors.New("addKeyIfNotFound was set to true but a variable already exists with key " + nextVar.Key)
		}

		tfVar := Var{Key: nextVar.Key, Value: cfg.NewValue, Hcl: false, Sensitive: cfg.SensitiveVariable}

		if !config.readOnly {
			UpdateVariable(cfg.Organization, cfg.Workspace, nextVar.ID, tfVar)
		}
		return fmt.Sprintf("Replaced the value of %s from %s to %s", nextVar.Key, oldValue, cfg.NewValue), nil
	}

	// At this point, we haven't found a match
	if cfg.AddKeyIfNotFound {
		tfVar := Var{Key: cfg.SearchString, Value: cfg.NewValue, Hcl: false, Sensitive: cfg.SensitiveVariable}

		if !config.readOnly {
			CreateVariable(cfg.Organization, cfg.Workspace, tfVar)
		}
		return fmt.Sprintf("Added variable %s = %s", cfg.SearchString, cfg.NewValue), nil
	}

	return "No match found and no variable added", nil
}

type OAuthTokens struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			CreatedAt           time.Time `json:"created-at"`
			ServiceProviderUser string    `json:"service-provider-user"`
			HasSSHKey           bool      `json:"has-ssh-key"`
		} `json:"attributes"`
		Relationships struct {
			OauthClient struct {
				Data struct {
					ID   string `json:"id"`
					Type string `json:"type"`
				} `json:"data"`
				Links struct {
					Related string `json:"related"`
				} `json:"links"`
			} `json:"oauth-client"`
		} `json:"relationships"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
}

func getVCSToken(vcsUsername, orgName string) (string, error) {
	url := fmt.Sprintf(baseURL+"/organizations/%s/oauth-tokens", orgName)
	resp, err := callAPI(http.MethodGet, url, "", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var oauthTokens OAuthTokens
	if err := json.NewDecoder(resp.Body).Decode(&oauthTokens); err != nil {
		return "", err
	}

	vcsTokenID := ""
	for _, nextToken := range oauthTokens.Data {
		if nextToken.Attributes.ServiceProviderUser == vcsUsername {
			vcsTokenID = nextToken.ID
			break
		}
	}

	return vcsTokenID, nil
}

// UpdateWorkspace updates one attribute of one or more Terraform Cloud workspaces.
func UpdateWorkspace(params WorkspaceUpdateParams) error {
	if err := validateUpdateWorkspaceParams(params); err != nil {
		return err
	}

	foundWs := FindWorkspaces(params.Organization, params.WorkspaceFilter)
	if len(foundWs) == 0 {
		return fmt.Errorf("no workspaces found matching the filter '%s'\n", params.WorkspaceFilter)
	}

	if config.readOnly {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
		_, _ = fmt.Fprintln(w, "organization:\t", params.Organization)
		_, _ = fmt.Fprintln(w, "workspace filter:\t", params.WorkspaceFilter)
		_, _ = fmt.Fprintln(w, "attribute:\t", params.Attribute)
		_, _ = fmt.Fprintln(w, "value:\t", params.Value)
		_ = w.Flush()
		fmt.Println("workspaces:")
		for _, val := range foundWs {
			fmt.Println("    " + val)
		}
		fmt.Printf("Found %d workspace(s)\n", len(foundWs))
	}

	jsonObj := gabs.Wrap(map[string]any{
		"data": map[string]any{
			"type": "workspace",
		},
	})
	if _, err := jsonObj.SetP(parseVal(params.Value), "data.attributes."+params.Attribute); err != nil {
		return fmt.Errorf("unable to process attribute for update: %s", err)
	}
	postData := jsonObj.String()

	if config.debug {
		fmt.Printf("request body:\n    %s\n", postData)
	}
	if config.readOnly {
		return nil
	}
	for id, name := range foundWs {
		url := fmt.Sprintf(baseURL+"/workspaces/%s", id)
		resp, err := callAPI(http.MethodPatch, url, postData, nil)
		if err != nil {
			return err
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		fmt.Printf("set '%s' to '%s' on workspace %s\n", params.Attribute, params.Value, name)
		if config.debug {
			fmt.Printf("response:\n    %s\n", bodyBytes)
		}
	}
	fmt.Printf("Updated %d workspace(s)\n", len(foundWs))
	return nil
}

func parseVal(value string) any {
	if value == "null" {
		return nil
	}
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return i
	}
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return value
}

func validateUpdateWorkspaceParams(params WorkspaceUpdateParams) error {
	if config.debug {
		fmt.Printf("params:\n    %#v\n", params)
	}

	if len(params.WorkspaceFilter) < 3 {
		return fmt.Errorf("workspace filter must be at least 3 characters, given: '%s'", params.WorkspaceFilter)
	}

	return nil
}

// FindWorkspaces uses the `search[name]` parameter to retrieve a list of workspaces in Terraform Cloud that
// match the workspaceFilter by the workspace name. The list is returned as a map with the ID in the key
// and the name in the value.
func FindWorkspaces(organization, workspaceFilter string) map[string]string {
	u := NewTfcUrl(fmt.Sprintf("/organizations/%s/workspaces", organization))
	u.SetParam(paramPageSize, strconv.Itoa(pageSize))
	u.SetParam(paramSearchName, workspaceFilter)

	var attributeData [][]string
	for page := 1; ; page++ {
		u.SetParam(paramPageNumber, strconv.Itoa(page))
		resp, err := callAPI(http.MethodGet, u.String(), "", nil)
		if err != nil {
			log.Fatalln(err)
		}
		ws := parseWorkspacePage(resp, []string{"id", "name"})
		attributeData = append(attributeData, ws...)
		if len(ws) < pageSize {
			break
		}
	}

	foundWs := map[string]string{}
	for _, ws := range attributeData {
		foundWs[ws[0]] = ws[1]
	}
	return foundWs
}

// GetWorkspaceAttributes returns a list of all workspaces in `organization` and the values of the attributes requested
// in the `attributes` list. The value of unrecognized attribute names will be returned as `null`.
func GetWorkspaceAttributes(organization string, attributes []string) ([][]string, error) {
	u := NewTfcUrl(fmt.Sprintf("/organizations/%s/workspaces", organization))
	u.SetParam(paramPageSize, strconv.Itoa(pageSize))

	var attributeData [][]string
	for page := 1; ; page++ {
		u.SetParam(paramPageNumber, strconv.Itoa(page))
		resp, err := callAPI(http.MethodGet, u.String(), "", nil)
		if err != nil {
			log.Fatalln(err)
		}

		ws := parseWorkspacePage(resp, attributes)
		attributeData = append(attributeData, ws...)
		if len(ws) < pageSize {
			break
		}
	}
	return attributeData, nil
}

func parseWorkspacePage(resp *http.Response, attributes []string) [][]string {
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic("failed to close response body: " + err.Error())
		}
	}()

	parsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		panic(err)
	}

	wsAttributes := parsed.Search("data", "*", "attributes").Children()
	attributeData := make([][]string, len(wsAttributes))
	for i, ws := range wsAttributes {
		attributeData[i] = make([]string, len(attributes))
		for j, a := range attributes {
			var v any
			if a == "id" {
				v = parsed.Path(fmt.Sprintf("data.%d.id", i)).Data()
			} else {
				v = ws.Path(a).Data()
			}
			attributeData[i][j] = fmt.Sprintf("%v", v)
		}
	}
	return attributeData
}

func GetWorkspaceByName(organizationName, workspaceName string) (Workspace, error) {
	u := NewTfcUrl(fmt.Sprintf("/organizations/%s/workspaces/%s", organizationName, workspaceName))

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return Workspace{}, err
	}

	var ws WorkspaceJSON
	if err := json.NewDecoder(resp.Body).Decode(&ws); err != nil {
		return Workspace{}, fmt.Errorf("json decode error: %s", err)
	}

	return ws.Data, nil
}

type VariableSet struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Name           string    `json:"name"`
		Description    string    `json:"description"`
		Global         bool      `json:"global"`
		UpdatedAt      time.Time `json:"updated-at"`
		VarCount       int       `json:"var-count"`
		WorkspaceCount int       `json:"workspace-count"`
	} `json:"attributes"`
	Relationships struct {
		Organization struct {
			Data struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"organization"`
		Vars struct {
			Data []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"vars"`
		Workspaces struct {
			Data []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"data"`
		} `json:"workspaces"`
	} `json:"relationships"`
}

func GetVariableSet(org, vsName string) (*VariableSet, error) {
	list, err := GetAllVariableSets(org)
	if err != nil {
		return nil, fmt.Errorf("error getting list of variable sets in org: %w", err)
	}
	for _, l := range list.Data {
		if l.Attributes.Name == vsName {
			found := l
			return &found, nil
		}
	}
	return nil, nil
}

type VariableSetList struct {
	Data  []VariableSet `json:"data"`
	Links struct {
		Self  string `json:"self"`
		First string `json:"first"`
		Prev  any    `json:"prev"`
		Next  any    `json:"next"`
		Last  string `json:"last"`
	} `json:"links"`
}

func GetAllVariableSets(organizationName string) (VariableSetList, error) {
	u := NewTfcUrl(fmt.Sprintf("/organizations/%s/varsets", organizationName))

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return VariableSetList{}, err
	}

	var variableSetList VariableSetList
	if err := json.NewDecoder(resp.Body).Decode(&variableSetList); err != nil {
		return VariableSetList{}, fmt.Errorf("unexpected content retrieving variable set list: %w", err)
	}

	return variableSetList, nil
}

// TODO: make a config struct for this call?
func ApplyVariableSet(varsetID string, workspaceIDs []string) error {
	u := NewTfcUrl(fmt.Sprintf("/varsets/%s/relationships/workspaces", varsetID))
	data := gabs.New()
	_, err := data.ArrayOfSize(len(workspaceIDs), "data")
	if err != nil {
		return fmt.Errorf("ArrayOfSize failed in ApplyVariableSet: %w", err)
	}
	for i, id := range workspaceIDs {
		if _, err := data.S("data").SetIndex(map[string]any{
			"type": "workspace",
			"id":   id,
		}, i); err != nil {
			return fmt.Errorf("SetIndex failed in ApplyVariableSet: %w", err)
		}
	}
	postData := data.String()

	if config.debug {
		fmt.Printf("request body:\n    %s\n", postData)
	}
	if config.readOnly {
		return nil
	}
	_, err = callAPI(http.MethodPost, u.String(), postData, nil)
	// TODO: need to look at response?
	return err
}

func copyVariableSetList(sourceWorkspaceID, destinationWorkspaceID string) error {
	sets, err := ListWorkspaceVariableSets(sourceWorkspaceID)
	if err != nil {
		return fmt.Errorf("copy variable sets: %w", err)
	}
	if err := ApplyVariableSetsToWorkspace(sets, destinationWorkspaceID); err != nil {
		return fmt.Errorf("copy variable sets: %w", err)
	}
	return nil
}

func ApplyVariableSetsToWorkspace(sets VariableSetList, workspaceID string) error {
	var failed []string
	var err error
	for _, set := range sets.Data {
		err = ApplyVariableSet(set.ID, []string{workspaceID})
		if err != nil {
			failed = append(failed, set.Attributes.Name)
		}
	}
	if len(failed) == len(sets.Data) {
		return fmt.Errorf("failed to apply variable sets: %w", err)
	}
	if len(failed) > 0 {
		return fmt.Errorf("failed to apply variable sets %s: %w", strings.Join(failed, ", "), err)
	}
	return nil
}

func ListWorkspaceVariableSets(workspaceID string) (VariableSetList, error) {
	u := NewTfcUrl(fmt.Sprintf("/workspaces/%s/varsets", workspaceID))

	resp, err := callAPI(http.MethodGet, u.String(), "", nil)
	if err != nil {
		return VariableSetList{}, err
	}

	var variableSetList VariableSetList
	if err := json.NewDecoder(resp.Body).Decode(&variableSetList); err != nil {
		return variableSetList, fmt.Errorf("unexpected content retrieving variable set list: %w", err)
	}

	return variableSetList, nil
}

func AddRemoteStateConsumers(workspaceID string, consumerIDs []string) error {
	u := NewTfcUrl(fmt.Sprintf("/workspaces/%s/relationships/remote-state-consumers", workspaceID))

	data := gabs.New()
	_, err := data.ArrayOfSize(len(consumerIDs), "data")
	if err != nil {
		return fmt.Errorf("ArrayOfSize failed in AddRemoteStateConsumers: %w", err)
	}
	for i, id := range consumerIDs {
		if _, err := data.S("data").SetIndex(map[string]any{
			"type": "workspaces",
			"id":   id,
		}, i); err != nil {
			return fmt.Errorf("SetIndex failed in AddRemoteStateConsumers: %w", err)
		}
	}
	postData := data.String()

	_, err = callAPI(http.MethodPost, u.String(), postData, nil)
	return err
}
