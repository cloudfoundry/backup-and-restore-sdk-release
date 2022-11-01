// Package manifest parses a BOSH release manifest to extract package names and dependencies
package manifest

import (
	"gopkg.in/yaml.v3"
	"log"
	"sort"
)

// Parse takes the contents of a Manifest.MF and parses out the packages and their dependencies
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
func (m Manifest) Ordered() (pkgs []string) {
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

// Dependents returns the inverted package dependency structure so that you
// can determine what relies on a package, as opposed to what a package relies on.
// The result is a map, where:
// - keys are package names
// - values are a slice packages that depend on the package named in the key
// Only packages with dependents are returned, so packages which are not
// relied on by any other package will not be keys in the map.
func (m Manifest) Dependents() map[string][]string {
	result := make(map[string][]string)

	for pkgName, dependencies := range m {
		for _, d := range dependencies {
			result[d] = append(result[d], pkgName)
		}
	}

	return result
}
