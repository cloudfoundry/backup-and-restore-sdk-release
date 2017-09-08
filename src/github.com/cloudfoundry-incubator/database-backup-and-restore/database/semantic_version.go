package database

import (
	"fmt"
	"strings"
)

type semanticVersion struct {
	major string
	minor string
	patch string
}

func (v semanticVersion) String() string {
	return strings.Join([]string{v.major, v.minor, v.patch}, ".")
}

func (v semanticVersion) MinorVersionMatches(v2 semanticVersion) bool {
	return v.major == v2.major && v.minor == v2.minor
}

func parseFromString(stringVersion string) (semanticVersion, error) {
	versionNums := strings.Split(stringVersion, ".")

	if len(versionNums) != 3 {
		return semanticVersion{}, fmt.Errorf(`can't parse semver "%s"`, stringVersion)
	}

	semVer := semanticVersion{
		major: versionNums[0],
		minor: versionNums[1],
		patch: versionNums[2],
	}
	return semVer, nil
}
