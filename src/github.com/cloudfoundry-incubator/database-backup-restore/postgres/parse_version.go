package postgres

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/version"
)

func ParseVersion(str string) (version.SemanticVersion, error) {
	trimmed := strings.TrimSpace(str)
	words := strings.Split(trimmed, " ")
	if len(words) < 2 {
		return version.SemanticVersion{}, fmt.Errorf(`invalid postgres version: "%s"`, str)
	}
	stringVersion := words[1]
	return version.ParseSemVerFromString(stringVersion)
}
