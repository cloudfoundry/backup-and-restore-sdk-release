package postgres

import (
	"fmt"

	"database-backup-restore/config"
	"database-backup-restore/runner"
)

func NewPostgresCommand(config config.ConnectionConfig, tempFolderManager config.TempFolderManager, cmd string) runner.Command {
	cmdArgs := []string{
		fmt.Sprintf("--username=%s", config.Username),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
	}

	env := map[string]string{
		"PGPASSWORD": config.Password,
	}

	if config.Tls != nil {
		caCertFileName, _ := tempFolderManager.WriteTempFile(config.Tls.Cert.Ca)
		env["PGSSLROOTCERT"] = caCertFileName

		if config.Tls.SkipHostVerify {
			env["PGSSLMODE"] = "verify-ca"
		} else {
			env["PGSSLMODE"] = "verify-full"
		}

		if config.Tls.Cert.Certificate != "" {
			clientCertFileName, _ := tempFolderManager.WriteTempFile(config.Tls.Cert.Certificate)
			env["PGSSLCERT"] = clientCertFileName
		}

		if config.Tls.Cert.PrivateKey != "" {
			clientKeyFileName, _ := tempFolderManager.WriteTempFile(config.Tls.Cert.PrivateKey)
			env["PGSSLKEY"] = clientKeyFileName
		}
	}

	return runner.NewCommand(cmd).WithParams(cmdArgs...).WithEnv(env)
}
