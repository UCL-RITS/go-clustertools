package groupmems

import (
	"github.com/wfd3/go-groups/src/group"
)

func GetMemberNames(g string) ([]string, error) {
	grp, err := group.Lookup(g)
	if err != nil {
		return []string{}, err
	}
	return grp.Members, nil
}
