package mysql

import (
	"log"
	"os/exec"
	"regexp"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type ClientVersionDetector struct {
	mysqldumpPath string
}

func NewMysqlDumpUtilityVersionDetector(mysqldumpPath string) ClientVersionDetector {
	return ClientVersionDetector{mysqldumpPath: mysqldumpPath}
}

func (d ClientVersionDetector) GetVersion() (version.SemanticVersion, error) {
	// sample output: "mysqldump  Ver 10.16 Distrib 10.1.22-MariaDB, for Linux (x86_64)"
	// /mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),/
	clientCmd := exec.Command(d.mysqldumpPath, "-V")

	return extractVersionUsingCommand(clientCmd, `^mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),`), nil
}
func extractVersionUsingCommand(cmd *exec.Cmd, pattern string) version.SemanticVersion {
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalln("Error running command.", err)
	}

	r := regexp.MustCompile(pattern)
	matches := r.FindSubmatch(stdout)
	if matches == nil {
		log.Fatalln("Could not determine version by using search pattern:", pattern)
	}

	versionString := matches[1]

	r = regexp.MustCompile(`(\d+).(\d+).(\S+)`)
	matches = r.FindSubmatch(versionString)
	if matches == nil {
		log.Fatalln("Could not determine version by using search pattern:", pattern)
	}

	return version.SemanticVersion{
		Major: string(matches[1]),
		Minor: string(matches[2]),
		Patch: string(matches[3]),
	}
}
