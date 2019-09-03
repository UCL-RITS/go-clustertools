package main

import (
	"fmt"
)

func ExampleGetArchiveListing() {
	fmt.Println(GetArchiveListing("test_files/tmp.tar.gz"))
	fmt.Println(GetArchiveListing("test_files/tmp.tar.bz2"))
	fmt.Println(GetArchiveListing("test_files/tmp.zip"))
	// Output:
	// [templates/ templates/shebang.txt templates/main.txt templates/package_description.txt templates/basic_header.txt]
	// [templates/ templates/shebang.txt templates/main.txt templates/package_description.txt templates/basic_header.txt]
	// [templates/ templates/basic_header.txt templates/main.txt templates/package_description.txt templates/shebang.txt]
}
