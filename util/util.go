package util

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// GetFirstLineOfFile gets the first line of a file that is not commented or blank
func GetFirstLineOfFile(file string) (string, error) {
	var line string

	f, err := os.Open(file)
	if err != nil {
		return line, err
	}
	defer f.Close()

	mComment := regexp.MustCompile(`^#`)
	mBlank := regexp.MustCompile(`^\s*$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if mComment.MatchString(text) || mBlank.MatchString(text) {
			continue
		}
		line = strings.TrimSpace(text)
		break
	}

	if err := scanner.Err(); err != nil {
		return line, err
	}

	return line, err
}

// IsExist returns true if the path exists
func IsExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		// exist
		return true
	}
	// not exist
	return false
}

// IsNotExist returns true if the path does not exist
func IsNotExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// not exist
		return true
	}
	// exist
	return false
}
