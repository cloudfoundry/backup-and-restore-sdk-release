package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParsePostgresVersion", func() {
	It("parses out 9.6 version", func() {
		Expect(ParsePostgresVersion(
			" PostgreSQL 9.6.3 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(semanticVersion{
			major: "9", minor: "6", patch: "3",
		}))
	})

	It("parses out 9.4 version", func() {
		Expect(ParsePostgresVersion(
			" PostgreSQL 9.4.9 on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(semanticVersion{
			major: "9", minor: "4", patch: "9",
		}))
	})

	It("fails if the input is blank", func() {
		_, err := ParsePostgresVersion("")
		Expect(err).To(MatchError(`invalid postgres version: ""`))
	})

	It("fails if there is no version specified after 'PostgreSQL'", func() {
		_, err := ParsePostgresVersion(" PostgreSQL on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit")
		Expect(err).To(HaveOccurred())
	})
})
