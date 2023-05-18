package adhelper

import (
	"errors"
	"fmt"
	"regexp"
)

// Returns a slice containing the usernames of members of an AD group.
func GetADGroupMembers(ldapOpts *LdapOpts, groupname string) ([]string, error) {

	expression := fmt.Sprintf("(&(objectCategory=Group)(cn=%s))", groupname)
	results, err := RunADSearch(ldapOpts, expression, []string{"member"})
	if err != nil {
		return nil, err
	}

	members := []string{}
	for _, entry := range results.Entries {
		for _, memberCN := range entry.GetAttributeValues("member") {
			membername, err := deCN(memberCN)
			if err != nil {
				return nil, err
			}
			members = append(members, membername)
		}
	}

	return members, nil
}

func GetADDeptMembers(ldapOpts *LdapOpts, deptName string) ([]string, error) {
	expression := fmt.Sprintf("(department=%s)", deptName)
	results, err := RunADSearch(ldapOpts, expression, []string{"cn"})
	if err != nil {
		return nil, err
	}

	members := []string{}
	for _, entry := range results.Entries {
		for _, memberCN := range entry.GetAttributeValues("cn") {
			membername, err := deCN(memberCN)
			if err != nil {
				return nil, err
			}
			members = append(members, membername)
		}
	}

	return members, nil
}

// In theory this should match any AD Canonical Name (CN) with a submatch of just the non-canonical name.
var CNre = regexp.MustCompile(`[Cc][Nn]=([A-Za-z0-9]+)(?:,[Oo][Uu]=[A-Za-z0-9-]+)*(?:,[Dd][Cc]=[A-Za-z0-9-]+)*`)

// Gets just the actual name from a full CN returned from AD.
func deCN(s string) (string, error) {
	matches := CNre.FindStringSubmatch(s)
	if matches == nil {
		return "", errors.New("attempted to deCN something that didn't seem to be a CN")
	}
	return matches[1], nil
}
