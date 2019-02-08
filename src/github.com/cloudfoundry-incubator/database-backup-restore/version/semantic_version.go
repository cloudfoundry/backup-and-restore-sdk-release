package version

import (
	"fmt"
	"regexp"
	"strings"
)

type DatabaseServerVersion struct {
	Implementation  string
	SemanticVersion SemanticVersion
}

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

func ParseSemVerFromString(stringVersion string) (SemanticVersion, error) {
	r := regexp.MustCompile(`(\d+).(\d+).(\S+)`)
	matches := r.FindSubmatch([]byte(stringVersion))
	if matches == nil {
		r = regexp.MustCompile(`(\d+).(\S+)`) //case where patch is omitted
		matches = r.FindSubmatch([]byte(stringVersion))

		if matches == nil {
			return SemanticVersion{}, fmt.Errorf(`could not parse semver: "%s"`, stringVersion)
		}

		matches = append(matches, []byte("0")) //patch is omitted, so we add patch 0
	}

	semVer := SemanticVersion{
		Major: string(matches[1]),
		Minor: string(matches[2]),
		Patch: string(matches[3]),
	}
	return semVer, nil
}

func SemVer(major, minor, patch string) SemanticVersion {
	return SemanticVersion{Major: major, Minor: minor, Patch: patch}
}
