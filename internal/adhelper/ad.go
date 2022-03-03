package adhelper

import (
	"crypto/tls"
	"fmt"

	"github.com/go-ldap/ldap/v3"
)

type LdapOpts struct {
	ServerUrl string `yaml:"server_url"`
	Username  string `yaml:"bind_username"`
	Password  string `yaml:"bind_password"`
	BaseDN    string `yaml:"base_dn"`
	Insecure  bool   `yaml:"allow_insecure"`
}

//var exampleLdapOpts = LdapOpts{
//	ServerUrl: "ldaps://ad.ucl.ac.uk/",
//	Username:  "AD\\user",
//	Password:  "hunter2",
//	BaseDN:    "DC=ad,DC=ucl,DC=ac,DC=uk",
//	Insecure:  true,
//}

func RunADSearch(opts *LdapOpts, searchExpression string, returnFields []string) (*ldap.SearchResult, error) {
	insecureDialOpt := ldap.DialWithTLSConfig(
		&tls.Config{
			InsecureSkipVerify: opts.Insecure,
		},
	)
	conn, err := ldap.DialURL(opts.ServerUrl, insecureDialOpt)
	if err != nil {
		return nil, fmt.Errorf("could not connect to LDAP server: %w", err)
	}
	defer conn.Close()
	err = conn.Bind(opts.Username, opts.Password)
	if err != nil {
		return nil, fmt.Errorf("could not bind on LDAP server: %w", err)
	}

	// Prototype, just for reference
	// func NewSearchRequest(
	//    BaseDN string,
	//    Scope, DerefAliases, SizeLimit, TimeLimit int,
	//    TypesOnly bool,
	//    Filter string,
	//    Attributes []string,
	//    Controls []Control,
	// ) *SearchRequest

	// This EscapeFilter prevents wildcard searches by escaping the asterisk.
	// We do want to be able to run wildcard searches.
	// TODO: make your own filter I guess
	// searchExpression = ldap.EscapeFilter(searchExpression))
	// In the meanwhile, this is... a little unsafe

	searchReq := ldap.NewSearchRequest(
		opts.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		searchExpression,
		returnFields, //empty attributes means all of them
		[]ldap.Control{},
	)

	sr, err := conn.SearchWithPaging(searchReq, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search LDAP: %w", err)
	}
	return sr, nil

	// E.g. debug print
	//for _, entry := range sr.Entries {
	//    fmt.Printf("%s: %v\n", entry.DN, entry.GetAttributeValue("cn"))
	//}

}
