package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func readPassword(possibleFilenames []string) string {
	for _, filename := range possibleFilenames {
		file, err := os.Open(filename)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			log.Fatal("Fatal error: file "+filename+" exists but could not be read: ", err)
		}
		defer file.Close()
		fileContents, err := ioutil.ReadAll(file)
		return strings.TrimSpace(string(fileContents))
	}
	log.Fatalf("No valid password file could be found from following files: " + strings.Join(possibleFilenames, " "))
	return ""
}
