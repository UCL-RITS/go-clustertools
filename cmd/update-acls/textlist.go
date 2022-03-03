package main

import (
	"os"
	"strings"
)

func GetTextListMembers(name string) ([]string, error) {
	fileBytes, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(fileBytes)), "\n"), nil
}

func SetTextListMembers(name string, list []string) error {
	outputString := strings.Join(list, "\n") + "\n"
	err := os.WriteFile(name, []byte(outputString), 0600)
	if err != nil {
		return err
	}
	return nil
}
