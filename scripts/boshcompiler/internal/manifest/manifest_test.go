package manifest_test

import (
	"boshcompiler/internal/manifest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manifest", func() {
	Context("Parse()", func() {
		It("can parse a manifest", func() {
			m := manifest.Parse(fakeManifest())
			Expect(m).To(BeEquivalentTo(map[string][]string{
				"gcs-blobstore-backup-restorer":         {"golang-1-linux"},
				"libopenssl1":                           {},
				"database-backup-restorer":              {"golang-1-linux"},
				"database-backup-restorer-boost":        {},
				"database-backup-restorer-mysql-5.6":    {"libopenssl1"},
				"database-backup-restorer-mysql-5.7":    {"database-backup-restorer-boost", "libopenssl1"},
				"database-backup-restorer-postgres-9.4": {},
				"s3-blobstore-backup-restorer":          {"golang-1-linux"},
				"azure-blobstore-backup-restorer":       {"golang-1-linux"},
				"database-backup-restorer-mariadb":      {"libpcre2", "libopenssl1"},
				"database-backup-restorer-mysql-8.0":    {},
				"golang-1-linux":                        {},
				"libpcre2":                              {},
			}))
		})
	})

	Context("Dependents()", func() {
		It("can list packages with their dependencies", func() {
			m := manifest.Parse(fakeManifest())
			Expect(m.Dependents()).To(SatisfyAll(
				HaveKeyWithValue("golang-1-linux", ConsistOf("database-backup-restorer", "azure-blobstore-backup-restorer", "gcs-blobstore-backup-restorer", "s3-blobstore-backup-restorer")),
				HaveKeyWithValue("libpcre2", ConsistOf("database-backup-restorer-mariadb")),
				HaveKeyWithValue("libopenssl1", ConsistOf("database-backup-restorer-mariadb", "database-backup-restorer-mysql-5.6", "database-backup-restorer-mysql-5.7")),
				HaveKeyWithValue("database-backup-restorer-boost", ConsistOf("database-backup-restorer-mysql-5.7")),
			))
		})
	})

	Context("Ordered()", func() {
		It("can sort packages with the most heavily depended on first", func() {
			m := manifest.Parse(fakeManifest())
			Expect(m.Ordered()).To(Equal([]string{
				"golang-1-linux",
				"libopenssl1",
				"database-backup-restorer-boost",
				"libpcre2",
				"azure-blobstore-backup-restorer",
				"database-backup-restorer",
				"database-backup-restorer-mariadb",
				"database-backup-restorer-mysql-5.6",
				"database-backup-restorer-mysql-5.7",
				"database-backup-restorer-mysql-8.0",
				"database-backup-restorer-postgres-9.4",
				"gcs-blobstore-backup-restorer",
				"s3-blobstore-backup-restorer",
			}))
		})
	})
})
