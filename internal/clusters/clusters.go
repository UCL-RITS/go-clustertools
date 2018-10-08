package clusters

import (
	"errors"
	"os"
	"regexp"
	"strings"
)

type Cluster struct {
	Name             string
	AccountingDBName string
	HostsRegex       string
}

// Old Legion is here for completeness and in case it needs to be queried.
// In practice, it probably won't be used in jobhist, but might come up elsewhere.
var clusters = &[]Cluster{
	{"myriad", "myriad_sgelogs", "^(?:login1[23]|node-[hij]00a-[0-9]{3})$"},
	{"legion", "sgelogs2", "^(?:login0[56789]|node-[l-qs-z][0-9]{2}[a-f]-[0-9]{3})$"},
	{"grace", "grace_sgelogs", "^(?:login0[12]|node-r99a-[0-9]{3})$"},
	{"thomas", "thomas_sgelogs", "^(?:login0[34]|node-k98[a-t]-[0-9]{3})$"},
	{"michael", "michael_sgelogs", "^(?:login1[01]|node-k10[a-i]-0[0-3][0-9]|util11)$"},
	{"old_legion", "sgelogs", "^$"},
}

func GetClusterFromHostname(hostname string) (*Cluster, error) {
	for _, cluster := range *clusters {
		if regexp.MustCompile(cluster.HostsRegex).MatchString(hostname) {
			return &cluster, nil
		}
	}
	return nil, errors.New("no matching cluster found for hostname " + hostname)
}

func GetLocalCluster() (*Cluster, error) {
	var hostname string
	var err error
	var cluster *Cluster

	hostname, err = os.Hostname()
	if err != nil {
		return nil, err
	}

	hostname = strings.SplitN(hostname, ".", 2)[0]
	cluster, err = GetClusterFromHostname(hostname)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func GetLocalClusterName() (string, error) {
	cluster, err := GetLocalCluster()
	if err != nil {
		return "", nil
	}
	return cluster.Name, nil
}
