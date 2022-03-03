package main

import (
	"fmt"

	"github.com/UCL-RITS/go-clustertools/internal/groupmems"
)

func main() {
	g, err := groupmems.GetMemberNames("staff")
	if err != nil {
		panic(err)
	}
	fmt.Println("Getting members for ccsprcop:")
	fmt.Println(g)
}
