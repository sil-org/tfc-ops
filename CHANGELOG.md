<a id="v4.0.2"></a>
# [Hide debug detail (v4.0.2)](https://github.com/sil-org/tfc-ops/releases/tag/v4.0.2) - 2024-11-08



[Changes][v4.0.2]


<a id="v4.0.1"></a>
# [v4.0.1](https://github.com/sil-org/tfc-ops/releases/tag/v4.0.1) - 2024-10-04

## Changelog
* 7ac4229 Merge pull request [#78](https://github.com/sil-org/tfc-ops/issues/78) from sil-org/release/v4.0.0
* 10f4354 update module path to v4



[Changes][v4.0.1]


<a id="v4.0.0"></a>
# [v4.0.0](https://github.com/sil-org/tfc-ops/releases/tag/v4.0.0) - 2024-10-02

### Added
- Use API token in dotfiles ([ITSE-1492](https://itse.youtrack.cloud/issue/ITSE-1492)) (Resolves [#54](https://github.com/sil-org/tfc-ops/issues/54))

### Changed
- CallApi returns an error instead of panic

### Removed
- Deleted commented out code
- Removed TFVars struct as it was a strict subset of the Vars struct.
- Removed deprecated `--dry-run-mode` flag

[Changes][v4.0.0]


<a id="v3.5.4"></a>
# [v3.5.4](https://github.com/sil-org/tfc-ops/releases/tag/v3.5.4) - 2024-02-05

### Fixed
- Fixed `workspaces clone` command to omit the VCS config parameters if the token is not specified.

[Changes][v3.5.4]


<a id="v3.5.3"></a>
# [fix goreleaser (v3.5.3)](https://github.com/sil-org/tfc-ops/releases/tag/v3.5.3) - 2023-08-31

### Fixed
- Removed `replacements` option removed in [goreleaser 1.19](https://github.com/goreleaser/goreleaser/releases/tag/v1.19.0)

[Changes][v3.5.3]


<a id="v3.5.2"></a>
# [Better error handling (v3.5.2)](https://github.com/sil-org/tfc-ops/releases/tag/v3.5.2) - 2023-08-31

### Fixed
- Better error handling in `GetVarsFromWorkspace` to indicate the cause of a 404 from the Terraform Cloud API. Addresses [#60](https://github.com/sil-org/tfc-ops/issues/60).

[Changes][v3.5.2]


<a id="v3.5.0"></a>
# [v3.5.0](https://github.com/sil-org/tfc-ops/releases/tag/v3.5.0) - 2023-06-20

### Added 
- New library function `CreateRunTrigger` to configure workspaces to trigger a run from another workspace.
- New library function `ListRunTriggers` to list configured run triggers for a given workspace.
- New library function `FindRunTrigger` to find a run trigger given the source and destination workspaces.

[Changes][v3.5.0]


<a id="v3.4.0"></a>
# [v3.4.0](https://github.com/sil-org/tfc-ops/releases/tag/v3.4.0) - 2023-06-19

### Added
- Added a new library function, `CreateRun`, to start a Run

[Changes][v3.4.0]


<a id="v3.3.0"></a>
# [v3.3.0](https://github.com/sil-org/tfc-ops/releases/tag/v3.3.0) - 2023-06-15

### Added
- Added new function AddRemoteStateConsumers
- Added new CLI command "workspaces consumers"

### Fixed
- Use the proper `/v3` reference for internal package references. It conveniently worked before correcting the module line in go.mod (v3.2.1).

[Changes][v3.3.0]


<a id="v3.2.1"></a>
# [v3.2.1](https://github.com/sil-org/tfc-ops/releases/tag/v3.2.1) - 2023-06-15

### Fixed
- Added the missing /v3 on the module line in go.mod

[Changes][v3.2.1]


<a id="v3.2.0"></a>
# [Include variable sets in workspace clone (v3.2.0)](https://github.com/sil-org/tfc-ops/releases/tag/v3.2.0) - 2023-06-15

### Added
- Added a workspace clone option to apply the source workspace's variable sets to the new workspace.

[Changes][v3.2.0]


<a id="3.2.0"></a>
# [3.2.0](https://github.com/sil-org/tfc-ops/releases/tag/3.2.0) - 2023-06-15

### Added
- Added a workspace clone option to apply the source workspace's variable sets to the new workspace.

[Changes][3.2.0]


<a id="3.1.2"></a>
# [3.1.2](https://github.com/sil-org/tfc-ops/releases/tag/3.1.2) - 2023-04-25

### Fixed
- Fix `variables add` command to exit with a non-zero error code when the requested variable already exists.
- Fix error message in `variables add` command to include line termination and correct variable interpolation.

[Changes][3.1.2]


<a id="3.1.1"></a>
# [3.1.1](https://github.com/sil-org/tfc-ops/releases/tag/3.1.1) - 2023-03-09

## Changelog
* fa67f00 Merge pull request [#58](https://github.com/sil-org/tfc-ops/issues/58) from sil-org/feature/upgrade-deps-for-security-alerts
* ab6abf4 Merge pull request [#59](https://github.com/sil-org/tfc-ops/issues/59) from sil-org/develop
* 1ffe978 upgrade dependencies due to security alerts



[Changes][3.1.1]


<a id="3.1.0"></a>
# [3.1.0](https://github.com/sil-org/tfc-ops/releases/tag/3.1.0) - 2023-03-08

### Added
- New command, `varsets apply`, applies a variable set to one or more workspaces.
- New command, `variables add`, adds a new variable. Fails if the variable exists.
- New command, `variables delete`, deletes a variable.
- Added `-w --workspace` flag as persistent on all `variables` commands.
- Added unit test flow in Github Actions.
- New flag `--csv`: output variables list in CSV format.

### Changed
- Changed "Terraform Enterprise" to "Terraform Cloud" in help output text.

### Fixed
- Internally, changed global variable `dryRunMode` to `readOnlyMode`.
- Internal refactoring to remove the API token from most function signatures.
- Escape quotes and newlines in CSV content.
- Removed internal references to API V2 since all API V1 support has already been removed.
- Dependency update.

### Deprecated
- Deprecated `--dry-run-mode` for `workspaces update` and `variables update` commands to alleviate confusion with `workspaces clone -d`. Going forward, use the equivalent `-r --read-only-mode` instead.

### Included
- [#48](https://github.com/sil-org/tfc-ops/issues/48) 
- [#49](https://github.com/sil-org/tfc-ops/issues/49) 
- [#51](https://github.com/sil-org/tfc-ops/issues/51) 
- [#52](https://github.com/sil-org/tfc-ops/issues/52) 
- [#53](https://github.com/sil-org/tfc-ops/issues/53) 


[Changes][3.1.0]


<a id="3.0.0"></a>
# [3.0.0](https://github.com/sil-org/tfc-ops/releases/tag/3.0.0) - 2022-03-03

### Added
- `workspaces update` and `workspaces list` now accept any attribute defined in the Terraform Cloud API. `id` is also accepted.
- Added `version` command.
### Removed
- Breaking change: attribute names that didn't exactly match the API have been removed. These were: `createdat`, `workingdirectory`, `terraformversion`, and `vcsrepo`. They are still accessible as `created-at`, `working-directory`, `terraform-version`, and `vcs-repo.identifier`.
### Fixed
- `workspaces update` now uses the `search[name]` API parameter rather than retrieving the full list of workspaces.

[Changes][3.0.0]


<a id="2.1.1"></a>
# [2.1.1](https://github.com/sil-org/tfc-ops/releases/tag/2.1.1) - 2022-02-07

### Fixed
- update README for changes made previously
- change from `master` to `main` branch

[Changes][2.1.1]


<a id="2.1.0"></a>
# [2.1.0](https://github.com/sil-org/tfc-ops/releases/tag/2.1.0) - 2022-02-07

### Added
- Added `created-at`, `structured-run-output-enabled`, `terraform-version`, `vcs-repo.display-identifier`, `vcs-repo-oauth-token-id` attributes to "workspaces list" command
- Added `structured-run-output`, `vcs-repo.oauth-token-id` to "workspaces update" command
### Deprecated
- Deprecated `createdat`, `workingdirectory`, `terraformversion`, `vcsrepo` on "workspaces list" command to use attribute names that exactly match the names in the Terraform API. This would make it easier to programmatically reference attribute names, and full support for all workspace attributes
### Fixed
- By including `vcs-repo.display-identifier` this also addresses issue [#29](https://github.com/sil-org/tfc-ops/issues/29).

[Changes][2.1.0]


<a id="2.0.3"></a>
# [2.0.3](https://github.com/sil-org/tfc-ops/releases/tag/2.0.3) - 2021-04-30

## Changelog

0a28080 Merge pull request [#37](https://github.com/sil-org/tfc-ops/issues/37) from sil-org/develop
3eab6d0 use Github actions to run goreleaser



[Changes][2.0.3]


<a id="2.0.2"></a>
# [2.0.2](https://github.com/sil-org/tfc-ops/releases/tag/2.0.2) - 2021-04-29

## Changelog

540714a Merge pull request [#36](https://github.com/sil-org/tfc-ops/issues/36) from sil-org/develop
6408e3c install goreleaser config
e415db3 print "found (n) workspace(s)" in dry run mode



[Changes][2.0.2]


<a id="2.0.1"></a>
# [Bugfix (2.0.1)](https://github.com/sil-org/tfc-ops/releases/tag/2.0.1) - 2021-04-26

- fixed bug in workspaces update command
- removed binaries from repo

[Changes][2.0.1]


<a id="2.0.0"></a>
# [Rename and restructured (2.0.0)](https://github.com/sil-org/tfc-ops/releases/tag/2.0.0) - 2021-04-23



[Changes][2.0.0]


<a id="1.0.0"></a>
# [Ready for production use (1.0.0)](https://github.com/sil-org/tfc-ops/releases/tag/1.0.0) - 2018-03-16

This version was used to migrate around 75 production environments successfully so we believe it is ready for others to use in migrating their production environments. 

[Changes][1.0.0]


[v4.0.2]: https://github.com/sil-org/tfc-ops/compare/v4.0.1...v4.0.2
[v4.0.1]: https://github.com/sil-org/tfc-ops/compare/v4.0.0...v4.0.1
[v4.0.0]: https://github.com/sil-org/tfc-ops/compare/v3.5.4...v4.0.0
[v3.5.4]: https://github.com/sil-org/tfc-ops/compare/v3.5.3...v3.5.4
[v3.5.3]: https://github.com/sil-org/tfc-ops/compare/v3.5.2...v3.5.3
[v3.5.2]: https://github.com/sil-org/tfc-ops/compare/v3.5.0...v3.5.2
[v3.5.0]: https://github.com/sil-org/tfc-ops/compare/v3.4.0...v3.5.0
[v3.4.0]: https://github.com/sil-org/tfc-ops/compare/v3.3.0...v3.4.0
[v3.3.0]: https://github.com/sil-org/tfc-ops/compare/v3.2.1...v3.3.0
[v3.2.1]: https://github.com/sil-org/tfc-ops/compare/v3.2.0...v3.2.1
[v3.2.0]: https://github.com/sil-org/tfc-ops/compare/3.2.0...v3.2.0
[3.2.0]: https://github.com/sil-org/tfc-ops/compare/3.1.2...3.2.0
[3.1.2]: https://github.com/sil-org/tfc-ops/compare/3.1.1...3.1.2
[3.1.1]: https://github.com/sil-org/tfc-ops/compare/3.1.0...3.1.1
[3.1.0]: https://github.com/sil-org/tfc-ops/compare/3.0.0...3.1.0
[3.0.0]: https://github.com/sil-org/tfc-ops/compare/2.1.1...3.0.0
[2.1.1]: https://github.com/sil-org/tfc-ops/compare/2.1.0...2.1.1
[2.1.0]: https://github.com/sil-org/tfc-ops/compare/2.0.3...2.1.0
[2.0.3]: https://github.com/sil-org/tfc-ops/compare/2.0.2...2.0.3
[2.0.2]: https://github.com/sil-org/tfc-ops/compare/2.0.1...2.0.2
[2.0.1]: https://github.com/sil-org/tfc-ops/compare/2.0.0...2.0.1
[2.0.0]: https://github.com/sil-org/tfc-ops/compare/1.0.0...2.0.0
[1.0.0]: https://github.com/sil-org/tfc-ops/tree/1.0.0

<!-- Generated by https://github.com/rhysd/changelog-from-release v3.8.0 -->
