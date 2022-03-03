package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/UCL-RITS/go-clustertools/internal/rc"
	"github.com/UCL-RITS/go-clustertools/internal/stringsets"
)

type sgeACL struct {
	Name            string
	Type            []string
	OverrideTickets int
	FunctionalShare int
	Entries         []string
}

func GetSGEACLList() ([]string, error) {
	stdout, _, err := rc.RunCommand("qconf", []string{"-sul"}, []string{})
	if err != nil {
		return nil, fmt.Errorf("could not get list of SGE ACLs: %w", err)
	}
	return strings.Split(stdout, "\n"), nil
}

func SGEACLExists(aclName string) (bool, error) {
	aclList, err := GetSGEACLList()
	if err != nil {
		return false, err
	}
	for _, v := range aclList {
		if v == aclName {
			return true, nil
		}
	}
	return false, nil
}

func GetSGEACLMembers(aclName string) ([]string, error) {
	stdout, _, err := rc.RunCommand("qconf", []string{"-su", aclName}, []string{"SGE_SINGLE_LINE=1"})
	if err != nil {
		return []string{}, err
	}
	acl, err := ParseACLFromText(stdout)
	if err != nil {
		return []string{}, err
	}
	return acl.Entries, nil
}

func ParseACLFromText(text string) (*sgeACL, error) {
	// From: sge_types(5):
	// An object name is a sequence of up to 512 ASCII printing characters
	//  except SPACE, "/", ":", "´", "\", "[", "]", "{", "}", "|", "(", ")",
	//  "@", "%", "," or the '"' character itself.
	//
	nameRE := regexp.MustCompile(`^name\s+(?P<name>[A-Za-z0-9.+<>?^=-]+)\s*$`)
	typeRE := regexp.MustCompile(`^type\s+(?P<type>ACL|DEPT|ACL,DEPT|DEPT,ACL)$`)
	fshareRE := regexp.MustCompile(`^fshare\s+(?P<fshare>[0-9]+)\s*$`)
	oticketRE := regexp.MustCompile(`^oticket\s+(?P<oticket>[0-9]+)\s*$`)
	entriesRE := regexp.MustCompile(`^entries\s+(?P<entries>(?:[a-z0-9]+,)*[a-z0-9]+|NONE)\s*$`)
	// ^-- This RE only matches the single-line format.
	//     I tried to match the multi-line format, but even though the RE looked
	//     right, it didn't seem to want to match multiple middle-lines (lines
	//     beginning with whitespace and ending with `, \`).
	//     That means we have to do some checks to make sure we haven't received
	//     the multi-line format, I guess.

	// While SGE always returns the fields in the same order, it doesn't state that this
	//  is a requirement for input ACLs, so while doing a line-by-line regex may seem
	//  unnecessary it seems easy and appropriate.

	// We want to return an error if any of these fields turn up more than once.
	var gotName, gotType, gotFshare, gotOticket, gotEntries bool

	lines := strings.Split(text, "\n")

	// This should take care of the multi-line thing, tbh.
	// With the duplicate field and all-field checks, anyway.
	// There's a blank line so you get 5 fields + 1 = 6 lines total
	if len(lines) != 6 {
		return nil, fmt.Errorf("attempting to parse SGE ACL, got wrong number of lines (expected 6, got %d)\n%+v", len(lines), lines)
	}

	acl := &sgeACL{}
	for _, v := range lines {
		if match := nameRE.FindStringSubmatch(v); match != nil {
			if gotName == true {
				return nil, errors.New("found duplicate field (name) while attempting to parse SGE ACL")
			}
			gotName = true
			acl.Name = match[1]
		} else if match := typeRE.FindStringSubmatch(v); match != nil {
			if gotType == true {
				return nil, errors.New("found duplicate field (type) while attempting to parse SGE ACL")
			}
			gotType = true
			acl.Type = strings.Split(match[1], ",")
		} else if match := fshareRE.FindStringSubmatch(v); match != nil {
			if gotFshare == true {
				return nil, errors.New("found duplicate field (fshare) while attempting to parse SGE ACL")
			}
			gotFshare = true
			functionalShare, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, errors.New("failed to parse SGE ACL field (fshare)")
			}
			acl.FunctionalShare = functionalShare
		} else if match := oticketRE.FindStringSubmatch(v); match != nil {
			if gotOticket == true {
				return nil, errors.New("found duplicate field (oticket) while attempting to parse SGE ACL")
			}
			gotOticket = true
			overrideTickets, err := strconv.Atoi(match[1])
			if err != nil {
				return nil, errors.New("failed to parse SGE ACL field (oticket)")
			}
			acl.OverrideTickets = overrideTickets
		} else if match := entriesRE.FindStringSubmatch(v); match != nil {
			if gotEntries == true {
				return nil, errors.New("found duplicate field (entries) while attempting to parse SGE ACL")
			}
			gotEntries = true
			// Leave it empty if SGE says NONE
			// I'm not sure what SGE would do if you actually had a user called NONE
			if match[1] != "NONE" {
				acl.Entries = strings.Split(match[1], ",")
			}
		} else {
			// A blank line is expected and fine
			if v != "" {
				return nil, errors.New("unmatchable data found while attempting to parse SGE ACL")
			}
		}
	}

	if gotName && gotType && gotOticket && gotFshare && gotEntries {
		return acl, nil
	}
	return nil, errors.New("could not find all required fields while parseing SGE ACL")
}

func SetSGEACLMembers(aclName string, memberList []string) error {
	err := AddSGEACLMembers(aclName, memberList)
	if err != nil {
		return fmt.Errorf("error while adding members to SGE ACL %s: %w", aclName, err)
	}

	updatedMembers, err := GetSGEACLMembers(aclName)
	if err != nil {
		return fmt.Errorf("error while retrieving members list after addition for SGE ACL %s: %w", aclName, err)
	}

	updatedMemberSet := stringsets.NewFromSlice(updatedMembers)
	intendedMemberSet := stringsets.NewFromSlice(memberList)
	membersToRemove := updatedMemberSet.Difference(intendedMemberSet).AsSlice()

	err = RemoveSGEACLMembers(aclName, membersToRemove)
	if err != nil {
		return fmt.Errorf("error while removing members from SGE ACL %s: %w", aclName, err)
	}

	// Consistency check
	updatedMembers, err = GetSGEACLMembers(aclName)
	if err != nil {
		return fmt.Errorf("error while retrieving final updated members list for SGE ACL %s: %w", aclName, err)
	}
	updatedMemberSet = stringsets.NewFromSlice(updatedMembers)
	if updatedMemberSet.Equals(intendedMemberSet) {
		return nil
	}
	return fmt.Errorf("contents of SGE ACL %s after set do not match intended contents", aclName)
}

func AddSGEACLMembers(aclName string, memberList []string) error {
	commaJoinedMemberList := strings.Join(memberList, ",")
	if commaJoinedMemberList == "" {
		return nil
	}
	stdout, _, err := rc.RunCommand("qconf", []string{"-au", commaJoinedMemberList, aclName}, []string{})
	fmt.Println(stdout)
	return err
}

func RemoveSGEACLMembers(aclName string, memberList []string) error {
	commaJoinedMemberList := strings.Join(memberList, ",")
	if commaJoinedMemberList == "" {
		return nil
	}
	stdout, _, err := rc.RunCommand("qconf", []string{"-du", commaJoinedMemberList, aclName}, []string{})
	fmt.Println(stdout)
	return err
}

func GetSGEACLMembers_Example() {
	members, err := GetSGEACLMembers("Open")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(members)
}
