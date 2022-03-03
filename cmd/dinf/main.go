package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/UCL-RITS/go-clustertools/internal/adhelper"
	"github.com/go-ldap/ldap/v3"
	"gopkg.in/alecthomas/kingpin.v2"
)

var defaultPasswordFiles = []string{"~/.adpw", "/shared/ucl/etc/adpw"}

var description = `
A program to quickly query Active Directory for user and group data.
`

var (
	app = kingpin.New("dinf", description)

	insecure    = app.Flag("insecure", "Insecurely ignore server certificate.").Short('i').Short('k').Bool()
	searchField = app.Flag("search-field", "Field to search on. (e.g. mail, memberOf, sn)").Short('f').Default("CN").String()

	ldapUrl       = app.Flag("server", "URL of the LDAP/AD server to connect to.").Short('s').Default("ldaps://ldap-auth-ad-slb.ucl.ac.uk:636/").String()
	bindUser      = app.Flag("user", "User to authenticate to the server with. (\"Bind\" user.)").Short('u').Default(`AD\sa-ritsldap01`).String()
	bindPassword  = app.Flag("password", "Password to authenticate to the server with. (\"Bind\" password.) If empty or not provided, files will be used.").Short('p').Default("").String()
	bindCredsFile = app.Flag("creds-file", "File to get bind credentials from. (Default: first of ~/.adpw /shared/ucl/etc/adpw)").Short('c').String()
	searchBase    = app.Flag("base", "Search base in the LDAP tree.").Short('b').Default("DC=ad,DC=ucl,DC=ac,DC=uk").String()
	returnFields  = app.Flag("output", "Command-separated fields to show in output. (Default: all)").Short('o').String()

	searchTerm = app.Arg("search_term", "Search term (required).").Required().Strings()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	var realPassword string

	if *bindPassword == "" {
		if *bindCredsFile != "" {
			realPassword = readPassword([]string{*bindCredsFile})
		} else {
			realPassword = readPassword(defaultPasswordFiles)
		}
	} else {
		realPassword = *bindPassword
	}

	ldapOpts := adhelper.LdapOpts{
		ServerUrl: *ldapUrl,
		Username:  *bindUser,
		Password:  realPassword,
		BaseDN:    *searchBase,
		Insecure:  *insecure,
	}

	splitReturnFields := strings.Split(*returnFields, ",")
	if (len(splitReturnFields) == 1) && (splitReturnFields[0] == "") {
		splitReturnFields = []string{}
	}

	searchExpression := fmt.Sprintf("(%s=%s)", *searchField, (*searchTerm)[0])

	result, err := adhelper.RunADSearch(&ldapOpts, searchExpression, splitReturnFields)
	if err != nil {
		log.Fatal(err)
	}
	PrettierPrint(result, 2)
}

// I abstracted these from the ldap package so I could alter the format.

// PrettierPrint outputs a human-readable description with indenting
func PrettierPrint(s *ldap.SearchResult, indent int) {
	for _, entry := range s.Entries {
		PrettierEPrint(entry, indent)
	}
}

// PrettierPrint outputs a human-readable description with indenting
func PrettierEAPrint(e *ldap.EntryAttribute, indent int) {
	for k, v := range e.Values {
		fmt.Printf("%s%s[%d]: %s\n", strings.Repeat(" ", indent), e.Name, k, v)
	}
	//fmt.Printf("%s%s: %#v\n", strings.Repeat(" ", indent), e.Name, e.Values)
}

// PrettierPrint outputs a human-readable description indenting
func PrettierEPrint(e *ldap.Entry, indent int) {
	fmt.Printf("%sDN: %s\n", strings.Repeat(" ", indent), e.DN)
	for _, attr := range e.Attributes {
		PrettierEAPrint(attr, indent+2)
	}
}
