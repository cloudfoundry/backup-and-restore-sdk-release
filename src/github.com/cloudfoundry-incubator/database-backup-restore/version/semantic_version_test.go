package version

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SemanticVersion", func() {
	Describe("String", func() {
		It("returns the string representation of the semver", func() {
			semver := SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "3",
			}

			Expect(semver.String()).To(Equal("1.2.3"))
		})
	})

	Describe("ParseFromString", func() {
		It("parses from string with 3 parts", func() {
			Expect(ParseFromString("1.2.3")).To(Equal(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "3",
			}))
		})

		It("parses even if the patch version contains text", func() {
			Expect(ParseFromString("1.2.3-MariaDB rest of stuff")).To(Equal(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "3-MariaDB",
			}))
		})

		It("fails if string has 2 parts", func() {
			_, err := ParseFromString("1.2")
			Expect(err).To(MatchError(`could not parse semver: "1.2"`))
		})
	})

	Describe("MinorVersionMatches", func() {
		It("returns true if the major + minor versions match", func() {
			Expect(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "0",
			}.MinorVersionMatches(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "9",
			})).To(BeTrue())
		})

		It("returns false if the major versions differ", func() {
			Expect(SemanticVersion{
				Major: "2",
				Minor: "2",
				Patch: "9",
			}.MinorVersionMatches(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "9",
			})).To(BeFalse())
		})

		It("returns false if the minor versions differ", func() {
			Expect(SemanticVersion{
				Major: "1",
				Minor: "2",
				Patch: "9",
			}.MinorVersionMatches(SemanticVersion{
				Major: "1",
				Minor: "3",
				Patch: "9",
			})).To(BeFalse())
		})
	})
})
