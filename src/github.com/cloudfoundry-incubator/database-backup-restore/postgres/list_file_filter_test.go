package postgres_test

import (
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListFileFilter", func() {
	It("Removes lines that include EXTENSION or SCHEMA", func() {
		listFile := []byte(`;
; Archive created at 2017-09-20 13:19:14 UTC
;     dbname: db1505905996
;     TOC Entries: 11
;     Compression: -1
;     Dump Version: 1.12-0
;     Format: CUSTOM
;     Integer: 4 bytes
;     Offset: 8 bytes
;     Dumped from database version: 9.6.3
;     Dumped by pg_dump version: 9.6.3
;
;
; Selected TOC Entries:
;
2132; 1262 16385 DATABASE - db1505905996 vcap
3; 2615 2200 SCHEMA - public vcap
2133; 0 0 COMMENT - SCHEMA public vcap
1; 3079 12393 EXTENSION - plpgsql
2134; 0 0 COMMENT - EXTENSION plpgsql
185; 1259 16398 TABLE public people test_user
186; 1259 16404 TABLE public places test_user
2126; 0 16398 TABLE DATA public people test_user
2127; 0 16404 TABLE DATA public places test_user`)

		filteredListFile := postgres.ListFileFilter(listFile)

		Expect(filteredListFile).To(Equal([]byte(`;
; Archive created at 2017-09-20 13:19:14 UTC
;     dbname: db1505905996
;     TOC Entries: 11
;     Compression: -1
;     Dump Version: 1.12-0
;     Format: CUSTOM
;     Integer: 4 bytes
;     Offset: 8 bytes
;     Dumped from database version: 9.6.3
;     Dumped by pg_dump version: 9.6.3
;
;
; Selected TOC Entries:
;
2132; 1262 16385 DATABASE - db1505905996 vcap
185; 1259 16398 TABLE public people test_user
186; 1259 16404 TABLE public places test_user
2126; 0 16398 TABLE DATA public people test_user
2127; 0 16404 TABLE DATA public places test_user`)))
	})

	It("is a no op if there are no EXTENSION or SCHEMA lines", func() {
		listFile := []byte(`;
; Archive created at 2017-09-20 13:19:14 UTC
;     dbname: db1505905996
;     TOC Entries: 11
;     Compression: -1
;     Dump Version: 1.12-0
;     Format: CUSTOM
;     Integer: 4 bytes
;     Offset: 8 bytes
;     Dumped from database version: 9.6.3
;     Dumped by pg_dump version: 9.6.3
;
;
; Selected TOC Entries:
;
2132; 1262 16385 DATABASE - db1505905996 vcap
`)

		filteredListFile := postgres.ListFileFilter(listFile)

		Expect(filteredListFile).To(Equal([]byte(`;
; Archive created at 2017-09-20 13:19:14 UTC
;     dbname: db1505905996
;     TOC Entries: 11
;     Compression: -1
;     Dump Version: 1.12-0
;     Format: CUSTOM
;     Integer: 4 bytes
;     Offset: 8 bytes
;     Dumped from database version: 9.6.3
;     Dumped by pg_dump version: 9.6.3
;
;
; Selected TOC Entries:
;
2132; 1262 16385 DATABASE - db1505905996 vcap
`)))
	})
})
