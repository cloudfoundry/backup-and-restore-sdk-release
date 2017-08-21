package database

import (
	"fmt"
	"strings"
)

func PostgresVersionParser(str string) (semanticVersion, error) {
	trimmed := strings.TrimSpace(str)
	words := strings.Split(trimmed, " ")
	if len(words) < 2 {
		return semanticVersion{}, fmt.Errorf(`invalid postgres version: "%s"`, str)
	}
	stringVersion := words[1]
	return parseFromString(stringVersion)
}
