package mysql

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type Backuper struct {
	config       config.ConnectionConfig
	backupBinary string
	clientBinary string
}

func NewBackuper(config config.ConnectionConfig, utilitiesConfig config.UtilitiesConfig) Backuper {
	return Backuper{
		config:       config,
		backupBinary: utilitiesConfig.Mysql.Dump,
		clientBinary: utilitiesConfig.Mysql.Client,
	}
}

func (b Backuper) Action(artifactFilePath string) error {
	cmdArgs := []string{
		"-v",
		"--single-transaction",
		"--skip-add-locks",
		"--user=" + b.config.Username,
		"--host=" + b.config.Host,
		fmt.Sprintf("--port=%d", b.config.Port),
		"--result-file=" + artifactFilePath,
		b.config.Database,
	}

	cmdArgs = append(cmdArgs, b.config.Tables...)

	_, _, err := runner.Run(b.backupBinary, cmdArgs, map[string]string{"MYSQL_PWD": b.config.Password})

	return err
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
