package version

import (
	"fmt"
	"strings"
)

type SemanticVersion struct {
	Major string
	Minor string
	Patch string
}

func (v SemanticVersion) String() string {
	return strings.Join([]string{v.Major, v.Minor, v.Patch}, ".")
}

func (v SemanticVersion) MinorVersionMatches(v2 SemanticVersion) bool {
	return v.Major == v2.Major && v.Minor == v2.Minor
}

func ParseFromString(stringVersion string) (SemanticVersion, error) {
	versionNums := strings.Split(stringVersion, ".")

	if len(versionNums) != 3 {
		return SemanticVersion{}, fmt.Errorf(`can't parse semver "%s"`, stringVersion)
	}

	semVer := SemanticVersion{
		Major: versionNums[0],
		Minor: versionNums[1],
		Patch: versionNums[2],
	}
	return semVer, nil
}

var V_9_4 = SemanticVersion{Major: "9", Minor: "4"}
var V_9_6 = SemanticVersion{Major: "9", Minor: "6"}
