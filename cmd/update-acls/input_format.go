package main

import (
	"github.com/UCL-RITS/go-clustertools/internal/adhelper"
)

type Config struct {
	ADOptions adhelper.LdapOpts `yaml:"ad_options"`
	Lists     []AEUList         `yaml:"lists"`
}

type ListSpec struct {
	Users         []string `yaml:"users"`
	TextListFiles []string `yaml:"text_list_files"`
	ADGroups      []string `yaml:"ad_groups"`
	UnixGroups    []string `yaml:"unix_groups"`
	Departments   []string `yaml:"departments"`
	SGEACLs       []string `yaml:"sge_acls"`
	GoldProjects  []string `yaml:"gold_projects"`
}

type DestinationList struct {
	TextListFiles []string `yaml:"text_list_files"`
	SGEACLs       []string `yaml:"sge_acls"`
	GoldProjects  []string `yaml:"gold_projects"`
	// Ideally AD groups would be here but we don't currently have
	//  write access to any of those.
	// It'd probably have to be a full CN, too.
}

// All-encompassing user list.
// All includes happen before all excludes -- let's not make anything too messy.
type AEUList struct {
	Name         string           `yaml:"name"`
	Description  string           `yaml:"description`
	Include      *ListSpec        `yaml:"include"`
	Exclude      *ListSpec        `yaml:"exclude"`
	Filter       *ListSpec        `yaml:"filter"` // The idea of this is that you take (include), remove (exclude), and then intersect the result with (filter).
	Destinations *DestinationList `yaml:"destinations"`
	BuiltList    []string         `yaml:"-"` // This is a place to stash the results of expansion. It doesn't get read or written during YAML parsing.
}
