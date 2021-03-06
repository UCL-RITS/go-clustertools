package clusters

import (
	"errors"
	"io/ioutil"
	"strings"
)

// Old Legion is here for completeness and in case it needs to be queried.
// In practice, it probably won't be used much.
var clusterAccountingDBs = map[string]string{
	"myriad":   "myriad_sgelogs",
	"legion":   "legion_sgelogs",
	"grace":    "grace_sgelogs",
	"thomas":   "thomas_sgelogs",
	"michael":  "michael_sgelogs",
	"kathleen": "kathleen_sgelogs",
	"young":    "young_sgelogs", // I hate this
	"legion1":  "sgelogs",
	"legion2":  "sgelogs2",
	"legion3":  "legion_sgelogs", // Currently a synonym for "legion"
}

// We used to use regexes to work out which cluster the current hostname was from,
//  but since the namespace collapse of Legion from login{05..09} to login{01..02},
//  hostname can no longer be reliably used to find cluster name.
func GetClusterNameFromSGEIdent() (string, error) {
	clusterNameBytes, err := ioutil.ReadFile("/opt/sge/default/common/cluster_name")
	if err != nil {
		return "", errors.New("could not get cluster name: " + err.Error())
	}

	// The ident file has a trailing newline
	clusterName := strings.TrimSuffix(string(clusterNameBytes), "\n")

	return clusterName, nil
}

// Wrapper function so that caller does not need to know what method is used.
func GetLocalClusterName() (string, error) {
	clusterName, err := GetClusterNameFromSGEIdent()
	if err != nil {
		return "", err
	}
	return clusterName, nil
}

func GetClusterAccountingDBName(clusterName string) (string, error) {
	dbName, wasPresent := clusterAccountingDBs[clusterName]
	if wasPresent == false {
		return "", errors.New("This cluster [" + clusterName + "] does not have a known associated accounting DB.")
	}
	return dbName, nil
}

func GetLocalClusterAccountingDBName() (string, error) {
	clusterName, err := GetLocalClusterName()
	if err != nil {
		return "", err
	}
	return GetClusterAccountingDBName(clusterName)
}
