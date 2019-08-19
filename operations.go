package main

import (
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"os"
)

func makeSet(elements []string) map[string]bool {
	setElements := map[string]bool{}

	// Initialize the map with elements passed by arguments
	for _, element := range elements {
		setElements[element] = true
	}
	return setElements
}

// The difference operation for the two sets
func difference(one map[string]bool, two map[string]bool) map[string]bool {
	differenceSet := map[string]bool{}

	for element := range one {
		if !two[element] {
			differenceSet[element] = true
		}
	}
	return differenceSet
}

// Reverse an array
func reverse(commits []*object.Commit) {
	i := 0
	j := len(commits) - 1
	for i < j {
		commits[i], commits[j] = commits[j], commits[i]
		i++
		j--
	}
}

// Checks whether a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
