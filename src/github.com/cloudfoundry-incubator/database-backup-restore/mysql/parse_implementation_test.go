package mysql_test

import (
	. "github.com/cloudfoundry-incubator/database-backup-restore/mysql"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseImplementation", func() {
	It("detects a mysql implementation", func() {
		implementation, err := ParseImplementation("MySQL Community Server (GPL)")
		Expect(err).NotTo(HaveOccurred())
		Expect(implementation).To(Equal("mysql"))
	})

	It("detects a mariadb implementation", func() {
		implementation, err := ParseImplementation("MariaDB Server")
		Expect(err).NotTo(HaveOccurred())
		Expect(implementation).To(Equal("mariadb"))
	})

	It("throws an error if it does not recognise the implementation", func() {
		_, err := ParseImplementation("Hot new mysql fork")
		Expect(err).To(MatchError("unsupported mysql database: Hot new mysql fork"))
	})
})
