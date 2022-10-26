// Package manifest parses a BOSH release manifest to extract package names and dependencies
package manifest

import (
	"gopkg.in/yaml.v3"
	"log"
	"sort"
)

func Parse(manifest []byte) Manifest {
	result := make(Manifest)

	var m struct {
		Packages []struct {
			Name         string   `yaml:"name"`
			Dependencies []string `yaml:"dependencies"`
		} `yaml:"packages"`
	}

	if err := yaml.Unmarshal(manifest, &m); err != nil {
		log.Fatal(err)
	}

	for _, p := range m.Packages {
		result[p.Name] = p.Dependencies
	}

	return result
}

type Manifest map[string][]string

// Ordered returns a list of packages weighting packages that have lots of dependents towards the start
func (m Manifest) Ordered() []string {
	var pkgs []string
	for pkgName := range m {
		pkgs = append(pkgs, pkgName)
	}

	dependents := m.Dependents()

	sort.Strings(pkgs) // alphabetical, for consistency
	sort.SliceStable(pkgs, func(a, b int) bool {
		return len(dependents[pkgs[b]]) < len(dependents[pkgs[a]])
	})

	return pkgs
}

func (m Manifest) Dependents() map[string][]string {
	result := make(map[string][]string)

	for pkgName, dependencies := range m {
		for _, d := range dependencies {
			result[d] = append(result[d], pkgName)
		}
	}

	return result
}
