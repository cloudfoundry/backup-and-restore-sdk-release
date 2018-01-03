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
	}
	var postgresServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var mariadbServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var interactorFactory = database.NewInteractorFactory(
		utilitiesConfig,
		postgresServerVersionDetector,
		mariadbServerVersionDetector)

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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "6", Patch: "1"}, nil)
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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "4", Patch: "1"}, nil)
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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}, nil)
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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "6", Patch: "1"}, nil)
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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "4", Patch: "1"}, nil)
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
					postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "9", Minor: "5", Patch: "1"}, nil)
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

				postgresServerVersionDetector.GetVersionReturns(version.SemanticVersion{}, fmt.Errorf("server version detection test error"))
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

			Context("when the version is detected as MariaDB 10.1.24", func() {
				BeforeEach(func() {
					mariadbServerVersionDetector.GetVersionReturns(version.SemanticVersion{Major: "10", Minor: "1", Patch: "24"}, nil)
				})

				It("builds a mysql.Backuper", func() {
					Expect(factoryError).NotTo(HaveOccurred())
					Expect(interactor).To(BeAssignableToTypeOf(mysql.Backuper{}))
				})
			})
		})

		Context("when the action is 'restore'", func() {
			BeforeEach(func() {
				action = "restore"
			})

			It("builds a mysql.Restorer", func() {
				Expect(factoryError).NotTo(HaveOccurred())
				Expect(interactor).To(BeAssignableToTypeOf(mysql.Restorer{}))
			})
		})
	})

	Context("when the configured adapter is not supported", func() {
		BeforeEach(func() {
			action = "backup"
			connectionConfig = config.ConnectionConfig{Adapter: "unsupported"}
		})

		It("fails", func() {
			Expect(factoryError).To(MatchError("unsupported adapter/action combination: unsupported/backup"))
		})
	})

	Context("when the action is not supported", func() {
		BeforeEach(func() {
			action = "unsupported"
			connectionConfig = config.ConnectionConfig{Adapter: "postgres"}
		})

		It("fails", func() {
			Expect(interactor).To(BeNil())
			Expect(factoryError).To(MatchError("unsupported adapter/action combination: postgres/unsupported"))
		})
	})
})
