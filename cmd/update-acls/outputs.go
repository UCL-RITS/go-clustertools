package main

import (
	"fmt"

	"github.com/UCL-RITS/go-clustertools/internal/goldops"
)

func WriteListDestinations(list AEUList) []error {
	if list.Destinations == nil {
		return []error{fmt.Errorf("list %s contained no destinations", list.Name)}
	}

	errs := []error{}

	for _, v := range list.Destinations.TextListFiles {
		err := SetTextListMembers(v, list.BuiltList)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, v := range list.Destinations.SGEACLs {
		err := SetSGEACLMembers(v, list.BuiltList)
		if err != nil {
			errs = append(errs, err)
		}
	}

	for _, v := range list.Destinations.GoldProjects {
		err := goldops.SetGoldProjectMembers(v, list.BuiltList)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}
