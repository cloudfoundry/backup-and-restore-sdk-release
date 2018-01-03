package mysql

import (
	"fmt"
	"log"

	"os/exec"

	"regexp"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/version"
)

type ServerVersionDetector struct {
	mysqlPath string
}

func NewServerVersionDetector(mysqlPath string) ServerVersionDetector {
	return ServerVersionDetector{mysqlPath: mysqlPath}
}

func (d ServerVersionDetector) GetVersion(config config.ConnectionConfig) (version.SemanticVersion, error) {
	clientCmd := exec.Command(d.mysqlPath,
		"--skip-column-names",
		"--silent",
		fmt.Sprintf("--user=%s", config.Username),
		fmt.Sprintf("--password=%s", config.Password),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		"--execute=SELECT VERSION()")

	semanticVersion := extractVersionUsingCommand(clientCmd, `(.+)`)

	log.Printf("MYSQL server version %v\n", semanticVersion)

	return semanticVersion, nil
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

	semanticVersion := version.SemanticVersion{
		Major: string(matches[1]),
		Minor: string(matches[2]),
		Patch: string(matches[3]),
	}

	return semanticVersion
}
