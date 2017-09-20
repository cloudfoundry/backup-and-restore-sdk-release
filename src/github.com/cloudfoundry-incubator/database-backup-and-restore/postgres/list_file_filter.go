package postgres

import "strings"

func ListFileFilter(bytes []byte) []byte {
	outputLines := []string{}
	lines := strings.Split(string(bytes), "\n")
	for _, line := range lines {
		if strings.Contains(line, " EXTENSION ") || strings.Contains(line, " SCHEMA ") {
			continue
		}
		outputLines = append(outputLines, line)
	}
	output := strings.Join(outputLines, "\n")
	return []byte(output)
}
