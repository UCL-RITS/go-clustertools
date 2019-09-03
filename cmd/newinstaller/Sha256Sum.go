package main

// This code is mostly from https://golang.org/pkg/crypto/sha256/#example_New_file

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"os"
)

func GetSha256SumString(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
