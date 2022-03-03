package main

import (
	"fmt"
	"log"
	"os"

	"github.com/UCL-RITS/go-clustertools/internal/adhelper"
	"github.com/UCL-RITS/go-clustertools/internal/goldops"
	"github.com/UCL-RITS/go-clustertools/internal/groupmems"
	"github.com/UCL-RITS/go-clustertools/internal/stringsets"
	"gopkg.in/yaml.v2"
)

func (c *Config) ExpandAEULists() ([]AEUList, []error) {
	results := []AEUList{}
	errs := []error{}
	for _, v := range c.Lists {
		// NB: v is a *copy*, not a reference
		if v.Include == nil {
			// Take a shortcut if the list entry specifies no sources
			v.BuiltList = []string{}
			continue
		}
		inclList, inclErrs := c.expandListSpec(v.Include)

		exclList := stringsets.New()
		exclErrs := []error{}
		if v.Exclude != nil {
			exclList, exclErrs = c.expandListSpec(v.Exclude)
			inclList.DifferenceUpdate(exclList)
		}

		v.BuiltList = inclList.AsSlice()
		results = append(results, v)
		errs = append(errs, inclErrs...)
		errs = append(errs, exclErrs...)
	}

	return results, errs
}

func (c *Config) expandListSpec(spec *ListSpec) (*stringsets.StringSet, []error) {
	var set *stringsets.StringSet
	if spec.Users != nil {
		set = stringsets.NewFromSlice(spec.Users)
	} else {
		set = stringsets.New()
	}

	// It would probably be possible to make these group functions a modular interface but I'm
	//  not up to it right now.
	// The destinations, too.

	// Keeping track of the source of each entry wouldn't be *impossible* but would require a
	//  significantly more complicated data structure, I think.

	errors := []error{}

	for _, v := range spec.TextListFiles {
		groupMembers, err := GetTextListMembers(v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding text list: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	for _, v := range spec.ADGroups {
		groupMembers, err := adhelper.GetADGroupMembers(&c.ADOptions, v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding AD Group: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	for _, v := range spec.UnixGroups {
		groupMembers, err := groupmems.GetMemberNames(v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding UNIX group: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	for _, v := range spec.Departments {
		// TODO: write GetADDeptMembers
		groupMembers, err := adhelper.GetADDeptMembers(&c.ADOptions, v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding AD Dept: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	for _, v := range spec.SGEACLs {
		groupMembers, err := GetSGEACLMembers(v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding SGE ACL: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	for _, v := range spec.GoldProjects {
		groupMembers, err := goldops.GetGoldProjectMembers(v)
		if err != nil {
			wrappedErr := fmt.Errorf("error while expanding Gold project: %w", err)
			errors = append(errors, wrappedErr)
		}
		set.AddSlice(groupMembers)
	}

	return set, errors
}

func parseConfig(file string) (*Config, error) {
	c := Config{}
	configBytes, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("could not read config file: %w", err)
	}
	err = yaml.Unmarshal(configBytes, &c)
	if err != nil {
		return nil, fmt.Errorf("could not parse YAML from config file: %w", err)
	}
	return &c, nil
}

func parseConfig_Example() {
	c, err := parseConfig("fixtures/example.conf")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", c)

	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- config dump:\n%s\n\n", string(d))

	emptyConfig, err := yaml.Marshal(Config{})
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- empty config dump:\n%s\n\n", string(emptyConfig))
	c.ExpandAEULists()
}
