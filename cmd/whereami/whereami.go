package main

import (
	"fmt"
	"github.com/UCL-RITS/go-clustertools/internal/clusters"
)

func main() {
	clusterName, err := clusters.GetLocalClusterName()
	if err != nil {
		panic(err)
	}
	fmt.Println(clusterName)
}
