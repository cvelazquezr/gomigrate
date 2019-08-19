package main

import (
	"bufio"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4"

	. "gopkg.in/src-d/go-git.v4/_examples"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// The Data structure of the pom configuration metadata files
type Data struct {
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"dependencies>dependency"`
}

// The Data2 is the other type of dependencies for the pom.xml file
type Data2 struct {
	Dependencies []struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"dependencyManagement>dependencies>dependency"`
}

// The MetadataSnapshot structure will store a snapshot of the
// metadata at a certain point in time
type MetadataSnapshot struct {
	Hash string
	Dependencies []string
}

//Function for checkout to a particular snapshot
func checkout(hash plumbing.Hash, workTree *git.Worktree) {
	_ = workTree.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(hash.String()),
	})
}

// This function parse the metadata file of a project and return the dependencies
// given the model of the metadata
func getDependencies(metadataFilepath string) []string {
	if !fileExists(metadataFilepath) {
		return nil
	}

	xmlFile, err := ioutil.ReadFile(metadataFilepath)

	if err != nil {
		fmt.Println("Error opening file: ", err)
		return nil
	}

	var data Data
	_ = xml.Unmarshal([]byte(xmlFile), &data)

	var data2 Data2
	var dependencies []string

	if len(data.Dependencies) == 0 {
		_ = xml.Unmarshal(xmlFile, &data2)

		for _, dependency := range data2.Dependencies {
			// Keep with the groupId and artifactId only to avoid including possible update rules
			dependencyIdentifier := dependency.GroupID + " " + dependency.ArtifactID
			dependencies = append(dependencies, dependencyIdentifier)
		}
	} else {
		for _, dependency := range data.Dependencies {
			dependencyIdentifier := dependency.GroupID + " " + dependency.ArtifactID
			dependencies = append(dependencies, dependencyIdentifier)
		}
	}

	return dependencies
}

func extractRules(pathRepository string) {
	r, _ := git.PlainOpen(pathRepository)
	ref, _ := r.Head()
	workTree, _ := r.Worktree()
	commit, _ := r.CommitObject(ref.Hash())

	metadataFile := "pom.xml"
	commitIter, _ := r.Log(&git.LogOptions{From: commit.Hash, MetadataName: &metadataFile})

	// Store the dependencies history
	dependenciesHistory := map[string]MetadataSnapshot{}

	// Saving the information about the commits and
	// reversing the order of the process
	commitsArray:= []*object.Commit{}

	fmt.Println("Checking for commits with changes in the metadata file ...")
	_ = commitIter.ForEach(func(c *object.Commit) error {
		commitsArray = append(commitsArray, c)
		return nil
	})
	reverse(commitsArray)

	fmt.Println("Analysing the commits ...")
	f, _ := os.OpenFile("migration_rules_candidates.txt", os.O_APPEND|os.O_WRONLY, 0644)
	// Analysing the commits
	for _, commit := range commitsArray {
		metadataModified := commit.MetadataModified

		checkout(commit.Hash, workTree)

		for _, metadataFilePath := range metadataModified  {
			// Analyse the dependencies of each metadata file

			dependencies := getDependencies(pathRepository + "/" + metadataFilePath)

			if metaSnapshot, ok := dependenciesHistory[metadataFilePath]; ok {
				// The metadata file has been previously modified
				previousDependencies := makeSet(metaSnapshot.Dependencies)
				currentDependencies := makeSet(dependencies)

				// Removed and added libraries in the Dependencies
				removedLibraries := difference(previousDependencies, currentDependencies)
				addedLibraries := difference(currentDependencies, previousDependencies)

				if len(removedLibraries) > 0 && len(addedLibraries) > 0 {
					// Make pairs for the candidates to migration rule
					fmt.Println("Writing possible rules")
					for libA := range removedLibraries {
						for libB := range addedLibraries {
							textToWrite := libA + " -> " + libB + "\n"
							_, err := f.WriteString(textToWrite)
							CheckIfError(err)

							fmt.Print(textToWrite)
						}
					}
				}
			}
			dependenciesHistory[metadataFilePath] = MetadataSnapshot{
				Hash:         commit.Hash.String(),
				Dependencies: dependencies,
			}
		}
	}
}

func main() {
	csvFile, _ := os.Open("repositories_existence.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	directory := "data/"
	url := "https://github.com/"

	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
		}
		repositoryName := strings.Split(line[0], "/")[1]
		repositoryPath := directory + repositoryName

		fmt.Println("Cloning " + repositoryName)
		_, errorClone := git.PlainClone(repositoryPath, false, &git.CloneOptions{
			URL:               url + line[0],
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		CheckIfError(errorClone)

		// Extract the possible rules of migration
		extractRules(repositoryPath)

		// Delete the folder of the project
		errorRemove := os.RemoveAll(repositoryPath)
		CheckIfError(errorRemove)
	}
}