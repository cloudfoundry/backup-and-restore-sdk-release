package database_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/database"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/database/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TableCheckingInteractor", func() {
	var tableChecker *fakes.FakeTableChecker
	var interactor *fakes.FakeInteractor
	var tableCheckingInteractor database.TableCheckingInteractor
	var returnError error
	var cfg config.ConnectionConfig
	artifactPath := "/artifact/file/path"

	BeforeEach(func() {
		tableChecker = new(fakes.FakeTableChecker)
		interactor = new(fakes.FakeInteractor)
	})

	JustBeforeEach(func() {
		tableCheckingInteractor =
			database.NewTableCheckingInteractor(cfg, tableChecker, interactor)
		returnError = tableCheckingInteractor.Action(artifactPath)
	})

	Context("when tables are specified", func() {
		BeforeEach(func() {
			cfg = config.ConnectionConfig{
				Tables: []string{"table1", "table2", "table3"},
			}
		})

		Context("when the tables exist", func() {
			BeforeEach(func() {
				tableChecker.FindMissingTablesReturns([]string{}, nil)
				interactor.ActionReturns(fmt.Errorf("test error"))
			})

			It("delegates to the wrapped interactor", func() {
				By("passing the right tables to the TableChecker", func() {
					Expect(tableChecker.FindMissingTablesArgsForCall(0)).To(Equal(cfg.Tables))
				})

				By("calling its Action method", func() {
					Expect(interactor.ActionCallCount()).To(Equal(1))
					Expect(interactor.ActionArgsForCall(0)).To(Equal(artifactPath))
				})

				By("returning its return value", func() {
					Expect(returnError).To(MatchError("test error"))
				})
			})
		})

		Context("when some tables don't exist", func() {
			BeforeEach(func() {
				tableChecker.FindMissingTablesReturns([]string{"table2", "table3"}, nil)
			})

			It("fails", func() {
				By("passing the right tables to the TableChecker", func() {
					Expect(tableChecker.FindMissingTablesArgsForCall(0)).To(Equal(cfg.Tables))
				})

				By("not calling its Action method", func() {
					Expect(interactor.ActionCallCount()).To(Equal(0))
				})

				By("returning an informative error", func() {
					Expect(returnError).To(MatchError("can't find specified table(s): table2, table3"))
				})
			})
		})

		Context("when find missing tables returns an error", func() {
			BeforeEach(func() {
				tableChecker.FindMissingTablesReturns(nil, fmt.Errorf("missing tables test error"))
			})

			It("fails", func() {
				By("passing the right tables to the TableChecker", func() {
					Expect(tableChecker.FindMissingTablesArgsForCall(0)).To(Equal(cfg.Tables))
				})

				By("not calling its Action method", func() {
					Expect(interactor.ActionCallCount()).To(Equal(0))
				})

				By("returning an informative error", func() {
					Expect(returnError).To(MatchError("missing tables test error"))
				})
			})
		})
	})

	Context("when no tables specified", func() {
		BeforeEach(func() {
			cfg = config.ConnectionConfig{
				Tables: nil,
			}
			interactor.ActionReturns(fmt.Errorf("test error"))
		})

		It("delegates to the wrapped interactor", func() {
			By("not calling the TableChecker", func() {
				Expect(tableChecker.FindMissingTablesCallCount()).To(Equal(0))
			})

			By("calling its Action method", func() {
				Expect(interactor.ActionCallCount()).To(Equal(1))
				Expect(interactor.ActionArgsForCall(0)).To(Equal(artifactPath))
			})

			By("returning its return value", func() {
				Expect(returnError).To(MatchError("test error"))
			})
		})
	})
})
