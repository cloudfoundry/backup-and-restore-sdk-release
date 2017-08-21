package database

import "strings"

func PostgresVersionParser(str string) (semanticVersion, error) {
	trimmed := strings.TrimSpace(str)
	words := strings.Split(trimmed, " ")
	stringVersion := words[1]
	return parseFromString(stringVersion)
}
