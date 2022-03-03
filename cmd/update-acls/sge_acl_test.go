package main

import "testing"

var goodACL = `name   anacl 
type    ACL
fshare  0
oticket 0
entries usera1,userb2,userc3
`

var badACL1 = `name  anacl
name secondname
fshare 0
oticket 0
entries usera1,userb2,userc3
`

func TestParseSGEACL(t *testing.T) {
	r, e := ParseACLFromText("")
	if (r != nil) || (e == nil) {
		t.Errorf("expected failure to parse, got success")
	}

	r, e = ParseACLFromText(badACL1)
	if (r != nil) || (e == nil) {
		t.Errorf("expected failure to parse, got success")
	}

	r, e = ParseACLFromText(goodACL)
	if e != nil {
		t.Errorf("expected successful parse, got error")
	}
	if (r.Name != "anacl") &&
		(r.Type[0] != "ACL") &&
		(len(r.Type) != 1) &&
		(r.FunctionalShare != 0) &&
		(r.OverrideTickets != 0) &&
		(len(r.Entries) != 3) &&
		(r.Entries[0] != "usera1") &&
		(r.Entries[1] != "userb2") &&
		(r.Entries[2] != "userc3") {
		t.Errorf("parsed good ACL but got wrong data")
	}
}
