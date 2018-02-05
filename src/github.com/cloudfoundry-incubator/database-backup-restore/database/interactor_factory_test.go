package database_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/database"
	"github.com/cloudfoundry-incubator/database-backup-restore/database/fakes"
	"github.com/cloudfoundry-incubator/database-backup-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-restore/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InteractorFactory", func() {
	var utilitiesConfig = config.UtilitiesConfig{
		Postgres94: config.UtilityPaths{
			Dump:    "pg_p4_dump",
			Restore: "pg_p4_restore",
			Client:  "pg_p4_client",
		},
		Postgres96: config.UtilityPaths{
			Dump:    "pg_p6_dump",
			Restore: "pg_p6_restore",
			Client:  "pg_p6_client",
		},
		Mariadb: config.UtilityPaths{
			Dump:    "mariadb_dump",
			Restore: "mariadb_restore",
			Client:  "mariadb_client",
		},
		Mysql55: config.UtilityPaths{
			Dump:    "mysql_55_dump",
			Restore: "mysql_55_restore",
			Client:  "mysql_55_client",
		},
		Mysql56: config.UtilityPaths{
			Dump:    "mysql_56_dump",
			Restore: "mysql_56_restore",
			Client:  "mysql_56_client",
		},
		Mysql57: config.UtilityPaths{
			Dump:    "mysql_57_dump",
			Restore: "mysql_57_restore",
			Client:  "mysql_57_client",
		},
	}
	var postgresServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var mysqlServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var interactorFactory = database.NewInteractorFactory(
		utilitiesConfig,
		postgresServerVersionDetector,
		mysqlServerVersionDetector)

	var action database.Action
	var connectionConfig config.ConnectionConfig

	var interactor database.Interactor
	var factoryError error

	JustBeforeEach(func() {
		interactor, factoryError = interactorFactory.Make(action, connectionConfig)
	})

	Context("when the configured adapter is postgres", func() {
		BeforeEach(func() {
			connectionConfig = config.ConnectionConfig{Adapter: "postgres"}
		})

		Context("when the action is 'backup'", func() {
			BeforeEach(func() {
				action = "backup"
			})

			Context("when the version is detected as 9.6", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "6", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(
						database.NewTableCheckingInteractor(connectionConfig,
							postgres.NewTableChecker(connectionConfig, "pg_p6_client"),
							postgres.NewBackuper(connectionConfig, "pg_p6_dump"),
						),
					))
				})
			})

			Context("when the version is detected as 9.4", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "4", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(
						database.NewTableCheckingInteractor(connectionConfig,
							postgres.NewTableChecker(connectionConfig, "pg_p4_client"),
							postgres.NewBackuper(connectionConfig, "pg_p4_dump"),
						),
					))
				})
			})

			Context("when the version is detected as 9.5", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}},
						nil)
				})

				It("fails to build database.TableCheckingInteractor", func() {
					Expect(factoryError).To(MatchError(ContainSubstring("unsupported version of postgres")))
				})
			})
		})

		Context("when the action is 'restore'", func() {
			BeforeEach(func() {
				action = "restore"
			})

			Context("when the version is detected as 9.6", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "6", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(interactor).To(Equal(
						postgres.NewRestorer(connectionConfig, "pg_p6_restore"),
					))
					Expect(factoryError).NotTo(HaveOccurred())
				})
			})

			Context("when the version is detected as 9.4", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "4", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(interactor).To(Equal(
						postgres.NewRestorer(connectionConfig, "pg_p4_restore"),
					))
					Expect(factoryError).NotTo(HaveOccurred())
				})
			})

			Context("when the version is detected as 9.5", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"postgres", version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}},
						nil)
				})

				It("fails to build database.TableCheckingInteractor", func() {
					Expect(factoryError).To(MatchError(ContainSubstring("unsupported version of postgres")))
				})
			})
		})

		Context("when the server version detection fails", func() {
			BeforeEach(func() {
				action = "backup"
				connectionConfig = config.ConnectionConfig{Adapter: "postgres"}

				postgresServerVersionDetector.GetVersionReturns(
					version.DatabaseServerVersion{}, fmt.Errorf("server version detection test error"))
			})

			It("fails", func() {
				Expect(interactor).To(BeNil())
				Expect(factoryError).To(MatchError("server version detection test error"))
			})
		})
	})

	Context("when the configured adapter is mysql", func() {
		BeforeEach(func() {
			connectionConfig = config.ConnectionConfig{Adapter: "mysql"}
		})

		Context("when the action is 'backup'", func() {
			BeforeEach(func() {
				action = "backup"
			})

			Context("when the version is detected as MariaDB 10.1.30", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"mariadb", version.SemanticVersion{Major: "10", Minor: "1", Patch: "30"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(connectionConfig, "mariadb_dump",
						mysql.NewLegacySSLOptionsProvider())))
				})
			})

			Context("when the version is detected as MySQL 5.5.57", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"mysql", version.SemanticVersion{Major: "5", Minor: "5", Patch: "57"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(connectionConfig, "mysql_55_dump", mysql.NewLegacySSLOptionsProvider())))
				})
			})

			Context("when the version is detected as MySQL 5.6.37", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"mysql", version.SemanticVersion{Major: "5", Minor: "6", Patch: "37"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(connectionConfig, "mysql_56_dump", mysql.NewLegacySSLOptionsProvider())))
				})
			})

			Context("when the version is detected as MySQL 5.7.19", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"mysql", version.SemanticVersion{Major: "5", Minor: "7", Patch: "19"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(connectionConfig, "mysql_57_dump", mysql.NewDefaultSSLProvider())))
				})
			})

			Context("when the version is detected as the not supported MariaDB 5.5.58", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{"mariadb", version.SemanticVersion{Major: "5", Minor: "5", Patch: "58"}}, nil)
				})

				It("errors", func() {
					Expect(factoryError).To(MatchError("unsupported version of mariadb: 5.5"))
				})
			})
		})

		Context("when the action is 'restore'", func() {
			BeforeEach(func() {
				action = "restore"
			})

			Context("when the version is detected as MariaDB 10.1.30", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mariadb",
							version.SemanticVersion{Major: "10", Minor: "1", Patch: "30"}}, nil)
				})

				It("builds a mysql.Restorer", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewRestorer(
						connectionConfig,
						"mariadb_restore",
						mysql.NewLegacySSLOptionsProvider()),
					))
				})
			})

			Context("when the version is detected as MySQL 5.5.57", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mysql",
							version.SemanticVersion{Major: "5", Minor: "5", Patch: "57"}}, nil)
				})

				It("builds a mysql.Restorer", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewRestorer(
						connectionConfig,
						"mysql_55_restore",
						mysql.NewLegacySSLOptionsProvider()),
					))
				})
			})

			Context("when the version is detected as MySQL 5.6.37", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mysql",
							version.SemanticVersion{Major: "5", Minor: "6", Patch: "37"}}, nil)
				})

				It("builds a mysql.Restorer", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewRestorer(
						connectionConfig,
						"mysql_56_restore",
						mysql.NewLegacySSLOptionsProvider()),
					))
				})
			})

			Context("when the version is detected as MySQL 5.7.19", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mysql",
							version.SemanticVersion{Major: "5", Minor: "7", Patch: "19"}}, nil)
				})

				It("builds a mysql.Restorer", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewRestorer(
						connectionConfig,
						"mysql_57_restore",
						mysql.NewDefaultSSLProvider(),
					)))
				})
			})

			Context("when the version is detected as the not supported MariaDB 5.5.58", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mariadb",
							version.SemanticVersion{Major: "5", Minor: "5", Patch: "58"}}, nil)
				})

				It("errors", func() {
					Expect(factoryError).To(MatchError("unsupported version of mariadb: 5.5"))
				})
			})
		})
	})

	Context("when the configured adapter is not supported", func() {
		BeforeEach(func() {
			action = "backup"
			connectionConfig = config.ConnectionConfig{Adapter: "unsupported"}
		})

		It("fails", func() {
			Expect(factoryError).To(MatchError(
				"unsupported adapter/action combination: unsupported/backup"))
		})
	})

	Context("when the action is not supported", func() {
		BeforeEach(func() {
			action = "unsupported"
			connectionConfig = config.ConnectionConfig{Adapter: "postgres"}
		})

		It("fails", func() {
			Expect(interactor).To(BeNil())
			Expect(factoryError).To(MatchError(
				"unsupported adapter/action combination: postgres/unsupported"))
		})
	})
})
