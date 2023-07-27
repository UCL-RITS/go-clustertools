package main

import "time"

type VaspUser struct {
	LatterNames   string
	FormerNames   string
	EmailAddress  string
	ValidToString string
	ValidToTime   *time.Time
	EntryKind     string
	// InternalID    string // I don't think we need this field (yet?): it seems to be an internal user ID and it's in the code for the delete user button
}

func (vu *VaspUser) ValidToTimeString() string {
	if vu.ValidToTime == nil {
		return "-"
	}
	return vu.ValidToTime.Format("2006-01-02")
}

func (vu *VaspUser) LicencedFor() []string {
	var licences []string

	// Unhelpfully, they don't directly tell us what the users are licensed for, only
	//  the date their licence is valid until, and we have to translate that into
	//  a 5/6 entitlement by comparing it to the cut-off date for VASP 6 access.
	// This cut-off date is 2021-04-01.
	vasp6CutOffDate := time.Date(2021, time.April, 1, 0, 0, 0, 0, time.UTC)

	licences = append(licences, "vasp5")

	if (vu.ValidToTime == nil) || (vu.ValidToTime.After(vasp6CutOffDate)) {
		licences = append(licences, "vasp6")
	}

	return licences
}
