package database_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/database"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/database/fakes"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("InteractorFactory", func() {
	var utilitiesConfig = config.UtilitiesConfig{}
	var postgresServerVersionDetector = new(fakes.FakeServerVersionDetector)
	var interactorFactory = database.NewInteractorFactory(utilitiesConfig, postgresServerVersionDetector)

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

			It("builds a database.TableCheckingInteractor", func() {
				Expect(interactor).To(BeAssignableToTypeOf(database.TableCheckingInteractor{}))
				Expect(factoryError).NotTo(HaveOccurred())
			})
		})

		Context("when the action is 'restore'", func() {
			BeforeEach(func() {
				action = "restore"
			})

			It("builds a postgres.Restorer", func() {
				Expect(interactor).To(BeAssignableToTypeOf(postgres.Restorer{}))
				Expect(factoryError).NotTo(HaveOccurred())
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

			It("builds a database.VersionSafeInteractor", func() {
				Expect(interactor).To(BeAssignableToTypeOf(database.VersionSafeInteractor{}))
				Expect(factoryError).NotTo(HaveOccurred())
			})
		})

		Context("when the action is 'restore'", func() {
			BeforeEach(func() {
				action = "restore"
			})

			It("builds a mysql.Restorer", func() {
				Expect(interactor).To(BeAssignableToTypeOf(mysql.Restorer{}))
				Expect(factoryError).NotTo(HaveOccurred())
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

	Context("when the postgres server version detection fails", func() {
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
