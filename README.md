# Terraform Cloud Ops Tool

This application can be helpful in making copies/clones of a workspace and bringing its variables over
to the new one. It can also be used for listing or updating workspace attributes and listing or
modifying variables in workspaces.

## Installation

There are three ways to download/install this script:

1. Download a pre-built binary for your operating system from the [Releases](https://github.com/sil-org/tfc-ops/releases) page.
2. If you're a Go developer and want to modify the source before running, clone this repo and run with `go run main.go ...`

## Configuration

To provide access to HCP Terraform (Terraform Cloud) run the `terraform login` command and follow the prompts. This
will store a short-lived token on your computer. tfc-ops uses this token to make API calls to HCP Terraform.

## Environment vars
- `TFC_OPS_DEBUG` - Set to `true` to enable debug output
- `ATLAS_TOKEN` - An HCP Terraform token can be set as an environment variable. Get this by going to
  https://app.terraform.io/app/settings/tokens and generating a new token. The recommended alternative is to use
  the `terraform login` command to request a short-lived token.
- `ATLAS_TOKEN_DESTINATION` - Only necessary if cloning to a new organization in TF Cloud.

## Cloning a TF Cloud Workspace
Examples.

Get help about the command.

```$ tfc-ops workspaces clone -h```

Clone just the workspace (no variables) to the same organization.

```$ tfc-ops workspaces clone -o=my-org -s=source-workspace -n=new-workspace```

Clone a workspace, its variables and its state to a different organization in TF Cloud.

Note: Sensitive variables will get a placeholder in the new workspace, the value of
which will need to be corrected manually.  Also, the environment variables from the source
workspace will need to be moved in the destination workspace from the normal variables section
down to the environment variables section (e.g. `CONFIRM_DESTROY`).

```
$ tfc-ops workspaces clone -c=true -t=true -o=org1 -p=org2 -d=true \
$   -s=source-workspace -n=destination-workspace -v=org2-vcs-token
```


## Getting a list of all TF Cloud Workspaces with some of their attributes 
Examples.

Get help about the command.

```$ tfc-ops workspaces list -h```

List the workspaces with at least one of their attributes/pieces of data.

```$ tfc-ops workspaces list -o=gtis -a=id,name,created-at,environment,working-directory,terraform-version,vcs-repo.identifier```

## Usage

### General Help
```text
$ tfc-ops -h
Perform TF Cloud operations, e.g. clone a workspace or manage variables within a workspace

Usage:
  tfc-ops [command]

Available Commands:
  help        Help about any command
  variables   Update or List variables
  varsets     Commands for Variable Sets
  version     Show tfc-ops version
  workspaces  Clone, List, or Update workspaces

Flags:
  -h, --help   help for tfc-ops

Use "tfc-ops [command] --help" for more information about a command.
```

### Workspaces Help
```text
$ tfc-ops workspaces
Top level command for describing or updating workspaces or cloning a workspace

Usage:
  tfc-ops workspaces [command]

Available Commands:
  clone       Clone a Workspace
  list        List Workspaces
  update      Update Workspaces

Flags:
  -h, --help                  help for workspaces
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")

Use "tfc-ops workspaces [command] --help" for more information about a command.
```


### Workspace Clone Help
```text
$ tfc-ops workspaces clone -h
Clone a Terraform Cloud Workspace

Usage:
  tfc-ops workspaces clone [flags]

Flags:
  -t, --copyState                     optional (e.g. "-t=true") whether to copy the state of the Source Workspace (only possible if copying to a new account).
  -c, --copyVariables                 optional (e.g. "-c=true") whether to copy the values of the Source Workspace variables.
  -d, --differentDestinationAccount   optional (e.g. "-d=true") whether to clone to a different TF account.
  -h, --help                          help for clone
  -p, --new-organization string       Name of the Destination Organization in Terraform Cloud
  -v, --new-vcs-token-id string       The new organization's VCS repo's oauth-token-id
  -n, --new-workspace string          required - Name of the new Workspace in Terraform Cloud
  -s, --source-workspace string       required - Name of the Source Workspace in Terraform Cloud

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

### Workspace Consumers Help
```text
$ tfc-ops workspaces consumers -h
Add to workspace remote state consumers. (Possible future capability: list, replace, delete)

Usage:
  tfc-ops workspaces consumers [flags]

Flags:
      --consumers string   required - List of remote state consumer workspaces, comma-separated
  -h, --help               help for consumers
  -w, --workspace string   required - Partial workspace name to search across all workspaces

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

### Workspace List Help

Any workspace attribute that can be read by the Terraform API can be retrieved
by the `workspaces list` command. See the [Terraform API Docs](https://www.terraform.io/cloud-docs/api-docs/workspaces#show-workspace)
for full details.

```text
$ tfc-ops workspaces list -h
Lists the TF workspaces with (some of) their attributes

Usage:
  tfc-ops workspaces list [flags]

Flags:
  -a, --attributes string   required - Workspace attributes to list, use Terraform Cloud API workspace attribute names
  -h, --help                help for list

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

### Workspace Update Help

Any workspace attribute that can be updated by the Terraform API can be updated 
by the `workspaces update` command. See the [Terraform API Docs](https://www.terraform.io/cloud-docs/api-docs/workspaces#request-body-1)
for full details.

```text
$ tfc-ops workspaces update -h
Updates an attribute of Terraform workspaces

Usage:
  tfc-ops workspaces update [flags]

Flags:
  -a, --attribute string   required - Workspace attribute to update, use Terraform Cloud API workspace attribute names
  -h, --help               help for update
  -v, --value string       required - Value
  -w, --workspace string   required - Partial workspace name to search across all workspaces

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

### Variables Help
```text
$ tfc-ops variables
Top level command to update or lists variables in all workspaces

Usage:
  tfc-ops variables [command]

Available Commands:
  add         Add new variable (if not already present)
  delete      Delete variable
  list        Report on variables
  update      Update/add a variable in a Workspace

Flags:
  -h, --help                  help for variables
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
  -w, --workspace string      Name of the Workspace in Terraform Cloud

Use "tfc-ops variables [command] --help" for more information about a command.
```

### Variables List Help
```text
$ tfc-ops variables -h
Show the values of variables with a key or value containing a certain string

Usage:
  tfc-ops variables list [flags]

Flags:
  -h, --help                    help for list
  -k, --key_contains string     required if value_contains is blank - string contained in the Terraform variable keys to report on
  -v, --value_contains string   required if key_contains is blank - string contained in the Terraform variable values to report on

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
  -w, --workspace string      Name of the Workspace in Terraform Cloud
```

### Variables Update Help
```text
$ tfc-ops variables update -h
Update or add a variable in a Terraform Cloud Workspace based on a complete case-insensitive match

Usage:
  tfc-ops variables update [flags]

Flags:
  -a, --add-key-if-not-found            optional (e.g. "-a=true") whether to add a new variable if a matching key is not found.
  -h, --help                            help for update
  -n, --new-variable-value string       required - The desired new value of the variable
  -v, --search-on-variable-value        optional (e.g. "-v=true") whether to do the search based on the value of the variables. (Must be false if add-key-if-not-found is true
  -x, --sensitive-variable              optional (e.g. "-x=true") make the variable sensitive.
  -s, --variable-search-string string   required - The string to match in the current variables (either in the Key or Value - see other flags)

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
  -w, --workspace string      Name of the Workspace in Terraform Cloud
```

### Variables Delete Help
```text
Delete variable in matching workspace having the specified key

Usage:
  tfc-ops variables delete [flags]

Flags:
  -h, --help         help for delete
  -k, --key string   required - Terraform variable key to delete, must match exactly

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
  -w, --workspace string      Name of the Workspace in Terraform Cloud
```

### Variables Add Help
```text
Add variable in matching workspace having the specified key

Usage:
  tfc-ops variables add [flags]

Flags:
  -h, --help           help for add
  -k, --key string     required - Terraform variable key
  -v, --value string   required - Terraform variable value

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
  -w, --workspace string      Name of the Workspace in Terraform Cloud
```

### Variable Sets Apply Help
```text
Apply an existing variable set to workspaces

Usage:
  tfc-ops varsets apply [flags]

Flags:
  -h, --help                      help for apply
  -s, --set string                required - Terraform variable set to add
  -w, --workspace string          Name of the Workspace in Terraform Cloud
      --workspace-filter string   Partial workspace name to search across all workspaces

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

### Variable Sets List Help
```text
List variable sets applied to a workspace

Usage:
  tfc-ops varsets list [flags]

Flags:
  -h, --help                      help for list
  -w, --workspace string          Name of the Workspace in Terraform Cloud
      --workspace-filter string   Partial workspace name to search across all workspaces

Global Flags:
  -o, --organization string   required - Name of Terraform Cloud Organization
  -r, --read-only-mode        read-only mode (e.g. "-r")
```

## License
tfc-ops is released under the Apache 2.0 license. See 
[LICENSE](https://github.com/sil-org/tfc-ops/blob/main/LICENSE)
