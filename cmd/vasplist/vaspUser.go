package main

import (
	"strings"
	"time"
)

type VaspUser struct {
	LatterNames   string
	FormerNames   string
	EmailAddress  string
	ValidToString string
	ValidToTime   *time.Time
	EntryKind     string
	// InternalID    string // I don't think we need this field (yet?): it seems to be an internal user ID and it's in the code for the delete user button
}

// EntryKinds:
//  - "HPC license member"
//  - "Standard license member"
//  - "Primary contact person"
//  - "License signatory"

func (vu *VaspUser) ValidToTimeString() string {
	if vu.ValidToTime == nil {
		return "-"
	}
	return vu.ValidToTime.Format("2006-01-02")
}

func (vu *VaspUser) LicensedForString() string {
	licenses := vu.LicensedFor()

	if len(licenses) == 0 {
		return "-"
	}
	return strings.Join(licenses, ",")
}

func (vu *VaspUser) IsLicensedFor(licenseName string) bool {
	licenses := vu.LicensedFor()
	for _, l := range licenses {
		if l == licenseName {
			return true
		}
	}
	return false
}

func (vu *VaspUser) LicensedFor() []string {
	var licenses []string

	// Unhelpfully, they don't directly tell us what users are licensed for, only
	//  the date their license is valid until, and we have to translate that into
	//  a 5/6 entitlement by comparing it to the cut-off date for VASP 6 access.

	// Validity dates are only relevant for HPC license members, not the other classes
	//  of entry.
	if vu.EntryKind != "HPC license member" {
		return []string{"vasp5", "vasp6"}
	}

	// If there's no validity date at all, their license is no longer valid.
	if vu.ValidToTime == nil {
		return []string{}
	}

	// We've eliminated everyone we don't have to check dates for, so then
	//  we can check whether the person should have access to VASP 5 or 5 and 6.
	// The official cut-off date is 2019-07-01. (Last checked: 2023-11-27)
	vasp6CutOffDate := time.Date(2019, time.July, 1, 0, 0, 0, 0, time.UTC)

	licenses = append(licenses, "vasp5")

	if vu.ValidToTime.After(vasp6CutOffDate) {
		licenses = append(licenses, "vasp6")
	}

	return licenses
}

func getLicensedList(vul *[]*VaspUser, licenseName string) *[]*VaspUser {
	resultList := []*VaspUser{}
	for _, vu := range *vul {
		if vu.IsLicensedFor(licenseName) {
			resultList = append(resultList, vu)
		}
	}
	return &resultList
}
