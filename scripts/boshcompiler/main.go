package main

import (
	"boshcompiler/internal/concurrently"
	"boshcompiler/internal/manifest"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	switch len(os.Args) {
	case 0, 1:
		log.Fatalln("Required arg: <release directory>")
	case 2:
		compileRelease(os.Args[1])
	default:
		log.Fatalln("Too many args")
	}
}

func compileRelease(releasePath string) {
	start := time.Now()
	mf := manifest.Parse(must(os.ReadFile(filepath.Join(releasePath, "release.MF"))))
	deps := mf.Dependents()
	workers := runtime.NumCPU() * 2

	log.Printf("Compiling %d packages with %d workers.", len(mf), workers)

	concurrently.Run(workers, mf.Ordered(), mf, func(pkgName string) {
		compilePackage(releasePath, pkgName, deps[pkgName])
	})

	log.Printf("All packages compiled [%s]\n", since(start))
}

func compilePackage(releasePath, pkgName string, deps []string) {
	start := time.Now()
	log.Printf("Package %q compile starting%s.\n", pkgName, dependents(deps))

	boshInstallTarget := filepath.Join("/var/vcap/packages", pkgName)
	mkdir(boshInstallTarget)

	cmd := exec.Command("bash", filepath.Join(releasePath, "packages", pkgName, "packaging"))
	cmd.Env = append(os.Environ(), fmt.Sprintf("BOSH_INSTALL_TARGET=%s", boshInstallTarget))
	cmd.Dir = filepath.Join(releasePath, "packages", pkgName)

	if output, err := cmd.CombinedOutput(); err != nil {
		log.Fatalf("Package %q compile failed [%s]: %s\n\n%s\n", pkgName, since(start), err, output)
	}

	log.Printf("Package %q compile finished [%s].\n", pkgName, since(start))
}

func since(start time.Time) string {
	return time.Since(start).Truncate(time.Second).String()
}

func mkdir(path string) {
	if err := os.MkdirAll(path, 0777); err != nil {
		log.Fatal(err)
	}
}

func must[A any](v A, err error) A {
	if err != nil {
		log.Fatal(err)
	}
	return v
}

func dependents(d []string) string {
	switch len(d) {
	case 0:
		return ""
	default:
		return fmt.Sprintf(" (dependents: %s)", strings.Join(d, ", "))
	}
}
