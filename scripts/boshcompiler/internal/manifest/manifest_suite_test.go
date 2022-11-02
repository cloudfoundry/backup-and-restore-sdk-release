package manifest_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestPkgs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifest Suite")
}

func fakeManifest() []byte {
	return []byte(`
packages:
- name: azure-blobstore-backup-restorer
  dependencies:
  - golang-1-linux
- name: database-backup-restorer
  dependencies:
  - golang-1-linux
- name: database-backup-restorer-boost
  dependencies: []
- name: database-backup-restorer-mariadb
  dependencies:
  - libpcre2
  - libopenssl1
- name: database-backup-restorer-mysql-5.6
  dependencies:
  - libopenssl1
- name: database-backup-restorer-mysql-5.7
  dependencies:
  - database-backup-restorer-boost
  - libopenssl1
- name: database-backup-restorer-mysql-8.0
  dependencies: []
- name: database-backup-restorer-postgres-9.4
  dependencies: []
- name: gcs-blobstore-backup-restorer
  dependencies:
  - golang-1-linux
- name: golang-1-linux
  dependencies: []
- name: libopenssl1
  dependencies: []
- name: libpcre2
  dependencies: []
- name: s3-blobstore-backup-restorer
  dependencies:
  - golang-1-linux
`)
}
