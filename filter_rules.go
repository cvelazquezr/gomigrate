package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"strconv"
)

func main() {
	migrationFile, _ := os.Open("migration_rules_candidates.txt")
	reader := csv.NewReader(bufio.NewReader(migrationFile))

	// Store the frequency of the pairs
	frequencyPair := map[string]int{}
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println(err)
		}

		if _, ok := frequencyPair[line[0]]; ok {
			frequencyPair[line[0]] += 1
		} else {
			frequencyPair[line[0]] = 1
		}
	}

	fmt.Println("Filtering out rules ...")

	// Filter out the rules with frequency lower than a threshold
	mostFrequent := map[string]int{}
	for pair, freq := range frequencyPair {
		if freq > 3 {
			mostFrequent[pair] = freq
		}
	}

	// Sort the elements
	highestFrequencies := map[string]string{}

	keys := make([]string, 0, len(mostFrequent))
	for k := range mostFrequent{
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Println("Extracting rules with higher values ...")

	// Extract from similar rules the ones with highest frequencies
	for _, k := range keys {
		keyParts := strings.Split(k, " -> ")
		antecessor := keyParts[0]

		if _, ok := highestFrequencies[antecessor]; ok {
			rulesParts := strings.Split(highestFrequencies[antecessor], " -> ")
			freq, _ := strconv.Atoi(rulesParts[2])

			if freq < mostFrequent[k] {
				highestFrequencies[antecessor] = k + " -> " + rulesParts[2]
			}
		} else {
			highestFrequencies[antecessor] = k + " -> " + strconv.Itoa(mostFrequent[k])
		}
	}
	fmt.Println("Writing the rules to a file ...")

	// Write to a file the final rules
	f, _ := os.OpenFile("migration_rules.txt", os.O_APPEND|os.O_WRONLY, 0644)

	for _, rule := range highestFrequencies {
		partsRule := strings.Split(rule, " -> ")
		textToWrite := partsRule[0] + " -> " + partsRule[1] + " : " + partsRule[2] + "\n"

		_, err := f.WriteString(textToWrite)

		if err != nil {
			return
		}
	}
}
