package postgres

import (
	"fmt"

	"io/ioutil"
	"log"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

func NewPostgresCommand(config config.ConnectionConfig, cmd string) runner.Command {
	cmdArgs := []string{
		fmt.Sprintf("--username=%s", config.Username),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
	}

	env := map[string]string{
		"PGPASSWORD": config.Password,
	}

	if config.Tls != nil {
		caCertFile, err := ioutil.TempFile("", "")
		if err != nil {
			log.Fatalf("error creating temp file: %s", err)
		}
		ioutil.WriteFile(caCertFile.Name(), []byte(config.Tls.Cert.Ca), 0777)
		env["PGSSLROOTCERT"] = caCertFile.Name()

		if config.Tls.SkipHostVerify {
			env["PGSSLMODE"] = "verify-ca"
		} else {
			env["PGSSLMODE"] = "verify-full"
		}

		if config.Tls.Cert.Certificate != "" {
			clientCertFile, err := ioutil.TempFile("", "")
			if err != nil {
				log.Fatalf("error creating temp file: %s", err)
			}
			ioutil.WriteFile(clientCertFile.Name(), []byte(config.Tls.Cert.Certificate), 0777)
			env["PGSSLCERT"] = clientCertFile.Name()
		}

		if config.Tls.Cert.PrivateKey != "" {
			clientKeyFile, err := ioutil.TempFile("", "")
			if err != nil {
				log.Fatalf("error creating temp file: %s", err)
			}
			ioutil.WriteFile(clientKeyFile.Name(), []byte(config.Tls.Cert.PrivateKey), 0777)
			env["PGSSLKEY"] = clientKeyFile.Name()
		}
	}

	return runner.NewCommand(cmd).WithParams(cmdArgs...).WithEnv(env)
}
