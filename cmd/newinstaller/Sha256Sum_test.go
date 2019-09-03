package main

import (
	"fmt"
)

func ExampleGetSha256SumString() {
	fmt.Println(GetSha256SumString("test_files/tmp.zip"))
	// Output:
	// 76ba6a9c7b495217bd7b2ddb3d1ccd7420dcab6c76ae679b00184aa80b4d299a 
}
