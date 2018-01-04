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

func (d ServerVersionDetector) GetVersion(config config.ConnectionConfig) (version.DatabaseServerVersion, error) {
	versionCmd := exec.Command(d.mysqlPath,
		"--skip-column-names",
		"--silent",
		fmt.Sprintf("--user=%s", config.Username),
		fmt.Sprintf("--password=%s", config.Password),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		"--execute=SELECT VERSION()")

	semanticVersion := extractVersionUsingCommand(versionCmd, `(.+)`)

	log.Printf("MYSQL server version %v\n", semanticVersion)

	versionCommentCmd := exec.Command(d.mysqlPath,
		"--skip-column-names",
		"--silent",
		fmt.Sprintf("--user=%s", config.Username),
		fmt.Sprintf("--password=%s", config.Password),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		"--execute=SELECT @@VERSION_COMMENT")

	stdout, err := versionCommentCmd.Output()
	if err != nil {
		return version.DatabaseServerVersion{}, err
	}

	implementation, err := ParseImplementation(string(stdout))
	if err != nil {
		return version.DatabaseServerVersion{}, err
	}

	return version.DatabaseServerVersion{
		Implementation:  implementation,
		SemanticVersion: semanticVersion,
	}, nil
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

	semVer, err := version.ParseFromString(string(versionString))
	if matches == nil {
		log.Fatalln(err)
	}

	return semVer
}
