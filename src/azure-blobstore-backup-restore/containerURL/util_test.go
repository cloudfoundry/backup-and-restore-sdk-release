package containerURL_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"azure-blobstore-backup-restore/containerURL"

	"net/url"
)

var _ = Describe("Our common ContainerURL utities", func() {
	var urlToSanitise *url.URL
	When("given a URL that contains no secrets", func() {
		BeforeEach(func() {
			var err error
			urlToSanitise, err = url.Parse("https://example.com/some/innocuous/path?snapshot=some-contents")
			Expect(err).NotTo(HaveOccurred())
		})
		It("returns the URL unchanged", func() {
			sanitisedURL := containerURL.Sanitise(urlToSanitise.String())
			Expect(sanitisedURL).To(Equal(urlToSanitise.String()))
		})
	})

	When("given a URL that contains potential secrets", func() {
		BeforeEach(func() {
			var err error
			urlToSanitise, err = url.Parse("https://example.com/some/suspicious/path?not-snapshot=possible-secrets")
			Expect(err).NotTo(HaveOccurred())
		})
		It("redacts the potential secret", func() {
			sanitisedURL := containerURL.Sanitise(urlToSanitise.String())
			Expect(sanitisedURL).To(Equal("https://example.com/some/suspicious/path?not-snapshot=REDACTED"))
		})
	})

	When("given a URL with lots of parameters of every type", func() {
		BeforeEach(func() {
			var err error
			urlToSanitise, err = url.Parse("https://example.com/some/suspicious/path?not-snapshot=possible-secrets&snapshot=safe&anotherthing=more-possible-secrets&snapshot=also-safe")
			Expect(err).NotTo(HaveOccurred())
		})
		It("redacts the potential secrets and not the 'snapshot' params", func() {
			sanitisedURL := containerURL.Sanitise(urlToSanitise.String())
			Expect(sanitisedURL).To(Equal("https://example.com/some/suspicious/path?anotherthing=REDACTED&not-snapshot=REDACTED&snapshot=safe&snapshot=also-safe"))
		})
	})

	When("given an invalid URL", func() {
		It("returns its input unchanged", func() {
			// This was written on 2024-02-26 for use in logging.
			// If it gets an invalid URL, we want to log that
			// invalid URL for debugging.
			invalidURL := ":/This is not a URL"
			Expect(containerURL.Sanitise(invalidURL)).To(Equal(invalidURL))
		})
	})
})
