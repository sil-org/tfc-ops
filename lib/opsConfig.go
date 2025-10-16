package lib

import (
	"reflect"
)

// OpsConfig represents one row of the plan.csv file's contents
type OpsConfig struct {
	SourceOrg        string
	SourceName       string
	NewOrg           string
	NewName          string
	TerraformVersion string
	RepoID           string
	Branch           string
	Directory        string
}

// AsArray returns the values of the OpsConfig attributes
func (o *OpsConfig) AsArray() []string {
	return []string{
		o.SourceOrg,
		o.SourceName,
		o.NewOrg,
		o.NewName,
		o.TerraformVersion,
		o.RepoID,
		o.Branch,
		o.Directory,
	}
}

func (o *OpsConfig) getColNames() []string {
	val := reflect.ValueOf(o).Elem()
	cols := []string{}

	for i := range val.NumField() {
		typeField := val.Type().Field(i)
		cols = append(cols, typeField.Name)
	}

	return cols
}
