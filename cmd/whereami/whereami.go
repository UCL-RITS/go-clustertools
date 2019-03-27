package main

import (
	"fmt"
	"github.com/UCL-RITS/go-clustertools/internal/clusters"
	"os"
)

func main() {
	clusterName, err := clusters.GetLocalClusterName()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
	}
	fmt.Println(clusterName)
}
