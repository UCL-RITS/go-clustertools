package main

import (
	"fmt"
)

// I'm still not 100% about this design. :/
func ExampleGetFoIsFromListing() {
	fmt.Println(GetFoIsFromListing([]string{"a.txt", "cake.c", "my_script.d", "a_fish", "bees/configure", "more_cake", "setup.py"}, "fake.zip"))
	// Output:
	// [{fake.zip cake.c 4} {fake.zip bees/configure 1} {fake.zip setup.py 7}]
}
