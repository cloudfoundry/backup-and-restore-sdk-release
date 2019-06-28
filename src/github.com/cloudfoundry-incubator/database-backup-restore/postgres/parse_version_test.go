package postgres

import (
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParsePostgresVersion", func() {
	It("parses out 9.6 version", func() {
		Expect(ParseVersion(
			" PostgreSQL 9.6.3 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(version.SemanticVersion{
			Major: "9", Minor: "6", Patch: "3",
		}))
	})

	It("parses out 9.4 version", func() {
		Expect(ParseVersion(
			" PostgreSQL 9.4.9 on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(version.SemanticVersion{
			Major: "9", Minor: "4", Patch: "9",
		}))
	})

	It("parses out 10.6 version", func() {
		Expect(ParseVersion(
			" PostgreSQL 10.6 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit"),
		).To(Equal(version.SemanticVersion{
			Major: "10", Minor: "6", Patch: "0",
		}))
	})

	It("parses out 11.x version", func() {
		Expect(ParseVersion(
			" PostgreSQL 11.1 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit"),
		).To(Equal(version.SemanticVersion{
			Major: "11", Minor: "1", Patch: "0",
		}))
	})

	It("fails if the input is blank", func() {
		_, err := ParseVersion("")
		Expect(err).To(MatchError(`invalid postgres version: ""`))
	})

	It("fails if there is no version specified after 'PostgreSQL'", func() {
		_, err := ParseVersion(" PostgreSQL on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit")
		Expect(err).To(MatchError(ContainSubstring("could not parse semver")))
	})
})
