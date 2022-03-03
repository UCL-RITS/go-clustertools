package goldops

import (
	"fmt"
	"strings"

	"github.com/UCL-RITS/go-clustertools/internal/rc"
	"github.com/UCL-RITS/go-clustertools/internal/stringsets"
)

func DoesGoldProjectExist(name string) (bool, error) {
	projects, err := GetListOfGoldProjects()
	if err != nil {
		return false, err
	}

	if stringsets.NewFromSlice(projects).Has(name) {
		return true, nil
	}
	return false, nil
}

func GetListOfGoldProjects() ([]string, error) {
	stdout, _, err := rc.RunCommand("glsproject", []string{"--quiet", "--show", "Name", "--raw"}, []string{})
	if err != nil {
		return nil, fmt.Errorf("could not list gold projects: %w", err)
	}

	return strings.Split(strings.TrimSpace(stdout), "\n"), nil
}

func GetGoldProjectMembers(name string) ([]string, error) {
	stdout, _, err := rc.RunCommand("glsproject", []string{"-quiet", "-show Users", "--raw", name}, []string{})
	if err != nil {
		return nil, fmt.Errorf("could not list members for gold project %s: %w", name, err)
	}
	return strings.Split(strings.TrimSpace(stdout), "\n"), nil
}

func SetGoldProjectMembers(name string, list []string) error {

	csNames := strings.Join(list, ",")

	_, _, err := rc.RunCommand("gchproject", []string{"-p", name, "--addUsers", csNames}, []string{})
	if err != nil {
		return fmt.Errorf("failure while attempting to add users to gold project %s: %w", name, err)
	}

	// This add-check-remove logic could get factored out if the targets had an abstracted interface
	// Though then I'd have to implement add and delete separately
	// That might be sensible anyway but w/e
	updatedMembers, err := GetGoldProjectMembers(name)
	if err != nil {
		return fmt.Errorf("failure while retrieving updated members list for gold project %s: %w", name, err)
	}

	updatedMemberSet := stringsets.NewFromSlice(updatedMembers)
	intendedMemberSet := stringsets.NewFromSlice(list)
	membersToRemove := updatedMemberSet.Difference(intendedMemberSet)

	csRemovals := strings.Join(membersToRemove.AsSlice(), ",")

	_, _, err = rc.RunCommand("gchproject", []string{"-p", name, "--delUsers", csRemovals}, []string{})
	if err != nil {
		return fmt.Errorf("failure while attempting to remove users from gold project %s: %w", name, err)
	}

	return nil
}

func GetListOfGoldProjects_Example() {
	projs, err := GetListOfGoldProjects()
	if err != nil {
		panic(err)
	}
	for _, v := range projs {
		fmt.Printf("'%s'\n", v)
	}
}
