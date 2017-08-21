package database

import "strings"

type semanticVersion struct {
	major string
	minor string
	patch string
}

func (v semanticVersion) String() string {
	return strings.Join([]string{v.major, v.minor, v.patch}, ".")
}
