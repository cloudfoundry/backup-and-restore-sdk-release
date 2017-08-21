package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgresVersionParser", func() {
	It("should parse out 9.6 version", func() {
		Expect(PostgresVersionParser(
			" PostgreSQL 9.6.3 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(semanticVersion{
			major: "9", minor: "6", patch: "3",
		}))
	})

	It("should parse out 9.4 version", func() {
		Expect(PostgresVersionParser(
			" PostgreSQL 9.4.9 on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit"),
		).To(Equal(semanticVersion{
			major: "9", minor: "4", patch: "9",
		}))
	})
})
