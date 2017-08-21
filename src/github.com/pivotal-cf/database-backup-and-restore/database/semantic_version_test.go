package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SemanticVersion", func() {
	Describe("String", func() {
		It("returns the string representation of the semver", func() {
			semver := semanticVersion{
				major: "1",
				minor: "2",
				patch: "3",
			}

			Expect(semver.String()).To(Equal("1.2.3"))
		})
	})

	Describe("parseFromString", func() {
		It("parses from string with 3 parts", func() {
			Expect(parseFromString("1.2.3")).To(Equal(semanticVersion{
				major: "1",
				minor: "2",
				patch: "3",
			}))
		})

		It("fails if string has 4 parts", func() {
			_, err := parseFromString("1.2.4.4")
			Expect(err).To(MatchError(`can't parse semver "1.2.4.4"`))
		})

		It("fails if string has 2 parts", func() {
			_, err := parseFromString("1.2")
			Expect(err).To(MatchError(`can't parse semver "1.2"`))
		})
	})

	Describe("MinorVersionMatches", func() {
		It("returns true if the major + minor versions match", func() {
			Expect(semanticVersion{
				major: "1",
				minor: "2",
				patch: "0",
			}.MinorVersionMatches(semanticVersion{
				major: "1",
				minor: "2",
				patch: "9",
			})).To(BeTrue())
		})

		It("returns false if the major versions differ", func() {
			Expect(semanticVersion{
				major: "2",
				minor: "2",
				patch: "9",
			}.MinorVersionMatches(semanticVersion{
				major: "1",
				minor: "2",
				patch: "9",
			})).To(BeFalse())
		})

		It("returns false if the minor versions differ", func() {
			Expect(semanticVersion{
				major: "1",
				minor: "2",
				patch: "9",
			}.MinorVersionMatches(semanticVersion{
				major: "1",
				minor: "3",
				patch: "9",
			})).To(BeFalse())
		})
	})
})
