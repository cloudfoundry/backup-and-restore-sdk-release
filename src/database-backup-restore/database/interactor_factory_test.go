package database_test

import (
	"database-backup-restore/config"
	"database-backup-restore/database"
	"database-backup-restore/database/fakes"
	"database-backup-restore/mysql"
	"database-backup-restore/postgres"
	"database-backup-restore/version"
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("InteractorFactory", func() {
	var utilitiesConfig config.UtilitiesConfig
	var postgresServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var mysqlServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var tempFolderManager, _ = config.NewTempFolderManager()
	var interactorFactory database.InteractorFactory

	var action database.Action
	var connectionConfig config.ConnectionConfig

	var interactor database.Interactor
	var factoryError error

	JustBeforeEach(func() {
		interactorFactory = database.NewInteractorFactory(
			utilitiesConfig,
			postgresServerVersionDetector,
			mysqlServerVersionDetector,
			tempFolderManager)

		interactor, factoryError = interactorFactory.Make(action, connectionConfig)
	})

	BeforeEach(func() {
		utilitiesConfig = config.UtilitiesConfig{
			Postgres10: config.UtilityPaths{
				Dump:    "pg_p_10_dump",
				Restore: "pg_p_10_restore",
				Client:  "pg_p_10_client",
			},
			Postgres11: config.UtilityPaths{
				Dump:    "pg_p_11_dump",
				Restore: "pg_p_11_restore",
				Client:  "pg_p_11_client",
			},
			Postgres13: config.UtilityPaths{
				Dump:    "pg_p_13_dump",
				Restore: "pg_p_13_restore",
				Client:  "pg_p_13_client",
			},
			Mariadb: config.UtilityPaths{
				Dump:    "mariadb_dump",
				Restore: "mariadb_restore",
				Client:  "mariadb_client",
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
			Mysql80: config.UtilityPaths{
				Dump:    "mysql_80_dump",
				Restore: "mysql_80_restore",
				Client:  "mysql_80_client",
			},
		}
	})

	Context("when the configured adapter is postgres", func() {
		BeforeEach(func() {
			connectionConfig = config.ConnectionConfig{Adapter: "postgres"}
		})

		Context("when the action is 'backup'", func() {
			BeforeEach(func() {
				action = "backup"
			})

			Context("when the version is detected as 10.6", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "10", Minor: "6", Patch: "0"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(
						database.NewTableCheckingInteractor(connectionConfig,
							postgres.NewTableChecker(connectionConfig, "pg_p_10_client"),
							postgres.NewBackuper(
								connectionConfig,
								tempFolderManager,
								"pg_p_10_dump",
							),
						),
					))
				})
			})

			Context("when the version is detected as 11", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "11", Minor: "1", Patch: "0"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(
						database.NewTableCheckingInteractor(connectionConfig,
							postgres.NewTableChecker(connectionConfig, "pg_p_11_client"),
							postgres.NewBackuper(
								connectionConfig,
								tempFolderManager,
								"pg_p_11_dump",
							),
						),
					))
				})
			})

			Context("when the version is detected as 13", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "13", Minor: "2", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(
						database.NewTableCheckingInteractor(connectionConfig,
							postgres.NewTableChecker(connectionConfig, "pg_p_13_client"),
							postgres.NewBackuper(
								connectionConfig,
								tempFolderManager,
								"pg_p_13_dump",
							),
						),
					))
				})
			})

			Context("when the version is detected as 9.5", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}},
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

			Context("when the version is detected as 10.6", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "10", Minor: "6", Patch: "0"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(interactor).To(Equal(
						postgres.NewRestorer(
							connectionConfig,
							tempFolderManager,
							"pg_p_10_restore",
						),
					))
					Expect(factoryError).NotTo(HaveOccurred())
				})
			})

			Context("when the version is detected as 13", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "13", Minor: "2", Patch: "1"}},
						nil)
				})

				It("builds a database.TableCheckingInteractor", func() {
					Expect(interactor).To(Equal(
						postgres.NewRestorer(
							connectionConfig,
							tempFolderManager,
							"pg_p_13_restore",
						),
					))
					Expect(factoryError).NotTo(HaveOccurred())
				})
			})

			Context("when the version is detected as 9.5", func() {
				BeforeEach(func() {
					postgresServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "postgres", SemanticVersion: version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}},
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

			Context("when the version is detected as MariaDB 10.*", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "mariadb", SemanticVersion: version.SemanticVersion{Major: "10", Minor: "3"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(
						connectionConfig,
						"mariadb_dump",
						mysql.NewLegacySSLOptionsProvider(tempFolderManager),
						mysql.NewEmptyAdditionalOptionsProvider(),
					)))
				})
			})

			Context("when the version is detected as MySQL 5.6.37", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{Implementation: "mysql", SemanticVersion: version.SemanticVersion{Major: "5", Minor: "6", Patch: "37"}}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(
						connectionConfig,
						"mysql_56_dump",
						mysql.NewLegacySSLOptionsProvider(tempFolderManager),
						mysql.NewPurgeGTIDOptionProvider(),
					)))
				})
			})

			Context("when the version is detected as MySQL 5.7.19", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							Implementation: "mysql",
							SemanticVersion: version.SemanticVersion{
								Major: "5",
								Minor: "7",
								Patch: "19"}},
						nil,
					)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewBackuper(
						connectionConfig,
						"mysql_57_dump",
						mysql.NewDefaultSSLProvider(tempFolderManager),
						mysql.NewPurgeGTIDOptionProvider(),
					)))
				})
			})

			Context("when the version is detected as MySQL 8.0.27", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							Implementation: "mysql",
							SemanticVersion: version.SemanticVersion{
								Major: "8",
								Minor: "0",
								Patch: "27"}},
						nil,
					)
				})

				Context("when MySQL 8 is supported", func() {
					BeforeEach(func() {
						tempfile, err := os.CreateTemp("", "fake_mysql_for_bbr_sdk")
						Expect(err).NotTo(HaveOccurred())
						utilitiesConfig.Mysql80.Client = tempfile.Name()
					})

					AfterEach(func() {
						os.Remove(utilitiesConfig.Mysql80.Client)
					})

					It("builds a mysql.Backuper", func() {
						Expect(factoryError).NotTo(HaveOccurred())
						Expect(interactor).To(Equal(mysql.NewBackuper(
							connectionConfig,
							"mysql_80_dump",
							mysql.NewDefaultSSLProvider(tempFolderManager),
							mysql.NewPurgeGTIDOptionProvider(),
						)))
					})
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

			Context("when the version is detected as MariaDB 10.*", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mariadb",
							version.SemanticVersion{Major: "10", Minor: "3"}}, nil)
				})

				It("builds a mysql.Restorer", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(Equal(mysql.NewRestorer(
						connectionConfig,
						"mariadb_restore",
						mysql.NewLegacySSLOptionsProvider(tempFolderManager)),
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
						mysql.NewLegacySSLOptionsProvider(tempFolderManager)),
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
						mysql.NewDefaultSSLProvider(tempFolderManager),
					)))
				})
			})

			Context("when the version is detected as MySQL 8.0.27", func() {
				BeforeEach(func() {
					mysqlServerVersionDetector.GetVersionReturns(
						version.DatabaseServerVersion{
							"mysql",
							version.SemanticVersion{Major: "8", Minor: "0", Patch: "27"}}, nil)
				})

				Context("when MySQL 8 is supported", func() {
					BeforeEach(func() {
						tempfile, err := os.CreateTemp("", "fake_mysql_for_bbr_sdk")
						Expect(err).NotTo(HaveOccurred())
						utilitiesConfig.Mysql80.Client = tempfile.Name()
					})

					AfterEach(func() {
						os.Remove(utilitiesConfig.Mysql80.Client)
					})

					It("builds a mysql.Restorer", func() {
						Expect(factoryError).NotTo(HaveOccurred())
						Expect(interactor).To(Equal(mysql.NewRestorer(
							connectionConfig,
							"mysql_80_restore",
							mysql.NewDefaultSSLProvider(tempFolderManager),
						)))
					})
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
