package database_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/database"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/database/fakes"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionSafeInteractor", func() {
	var wrappedInteractor *fakes.FakeInteractor
	var serverVersionDetector *fakes.FakeServerVersionDetector
	var dumpUtilityVersionDetector *fakes.FakeDumpUtilityVersionDetector
	var cfg config.ConnectionConfig

	var versionSafeInteractor database.VersionSafeInteractor

	BeforeEach(func() {
		wrappedInteractor = new(fakes.FakeInteractor)
		serverVersionDetector = new(fakes.FakeServerVersionDetector)
		dumpUtilityVersionDetector = new(fakes.FakeDumpUtilityVersionDetector)
		cfg = config.ConnectionConfig{
			Username: "username",
			Password: "password",
		}

		versionSafeInteractor = database.NewVersionSafeInteractor(
			wrappedInteractor,
			serverVersionDetector,
			dumpUtilityVersionDetector,
			cfg,
		)
	})

	Context("when the server and dump utility versions match", func() {
		BeforeEach(func() {
			serverVersionDetector.GetVersionReturns(version.ParseFromString("1.2.3"))
			dumpUtilityVersionDetector.GetVersionReturns(version.ParseFromString("1.2.4"))
		})

		It("delegates to the wrapped interactor", func() {
			wrappedInteractor.ActionReturns(fmt.Errorf("the wrapped interactor has failed"))

			err := versionSafeInteractor.Action("artifact/file/path")

			By("passing the right config to the server version detector")
			Expect(serverVersionDetector.GetVersionArgsForCall(0)).To(Equal(cfg))

			By("calling its Action method")
			Expect(wrappedInteractor.ActionCallCount()).To(Equal(1))
			Expect(wrappedInteractor.ActionArgsForCall(0)).To(Equal("artifact/file/path"))

			By("returning its error")
			Expect(err).To(MatchError("the wrapped interactor has failed"))
		})
	})

	Context("when the server and dump utility versions don't match", func() {
		BeforeEach(func() {
			serverVersionDetector.GetVersionReturns(version.ParseFromString("1.2.3"))
			dumpUtilityVersionDetector.GetVersionReturns(version.ParseFromString("3.2.1"))
		})

		It("fails", func() {
			err := versionSafeInteractor.Action("artifact/file/path")

			By("passing the right config to the server version detector")
			Expect(serverVersionDetector.GetVersionArgsForCall(0)).To(Equal(cfg))

			By("not calling the wrapped interactor")
			Expect(wrappedInteractor.ActionCallCount()).To(Equal(0))

			By("returning an error")
			Expect(err).To(HaveOccurred())
		})
	})
})
