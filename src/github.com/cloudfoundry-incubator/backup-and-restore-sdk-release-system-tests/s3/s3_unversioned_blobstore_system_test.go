// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package s3_test

import (
	"time"

	"strconv"

	"io/ioutil"

	"os"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("S3 unversioned backup and restore", func() {
	var region string
	var bucket string
	var backupRegion string
	var backupBucket string
	var instanceArtifactDirPath string

	var localArtifact *os.File
	var backuperInstance JobInstance

	BeforeEach(func() {
		backuperInstance = JobInstance{
			Deployment:          MustHaveEnv("BOSH_DEPLOYMENT"),
			Name:                "s3-unversioned-backuper",
			Index:               "0",
			CommandOutputWriter: GinkgoWriter,
		}

		region = MustHaveEnv("S3_UNVERSIONED_BUCKET_REGION")
		bucket = MustHaveEnv("S3_UNVERSIONED_BUCKET_NAME")

		backupRegion = MustHaveEnv("S3_UNVERSIONED_BACKUP_BUCKET_REGION")
		backupBucket = MustHaveEnv("S3_UNVERSIONED_BACKUP_BUCKET_NAME")

		DeleteAllFilesFromBucket(region, bucket)
		DeleteAllFilesFromBucket(backupRegion, backupBucket)

		instanceArtifactDirPath = "/tmp/s3-unversioned-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		Expect(backuperInstance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)).To(Succeed())
		var err error
		localArtifact, err = ioutil.TempFile("", "blobstore-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(backuperInstance.RunSuccessfully("rm -rf " + instanceArtifactDirPath)).To(Succeed())
		err := os.Remove(localArtifact.Name())
		Expect(err).NotTo(HaveOccurred())
		DeleteAllFilesFromBucket(region, bucket)
		DeleteAllFilesFromBucket(backupRegion, backupBucket)
	})

	It("backs up and restores an unversioned bucket", func() {
		var (
			preBackupFiles   []string
			backupFiles      []string
			postRestoreFiles []string
		)

		By("backing up from the source bucket to the backup bucket", func() {
			WriteFileInBucket(region, bucket, "original/path/to/file", "FILE1")
			preBackupFiles = ListFilesFromBucket(backupRegion, backupBucket)

			Expect(backuperInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
				" /var/vcap/jobs/s3-unversioned-blobstore-backup-restorer/bin/bbr/backup")).To(Succeed())

			backupFiles = ListFilesFromBucket(backupRegion, backupBucket)
			Expect(backupFiles).To(ConsistOf(MatchRegexp(
				"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/my_bucket/original/path/to/file" + "$")))
			Expect(GetFileContentsFromBucket(backupRegion, backupBucket, backupFiles[0])).To(Equal("FILE1"))
		})

		By("writing a helpful backup artifact file", func() {
			session, err := backuperInstance.Download(
				instanceArtifactDirPath+"/blobstore.json", localArtifact.Name())

			Expect(err).NotTo(HaveOccurred())
			Expect(session).Should(gexec.Exit(0))

			fileContents, err := ioutil.ReadFile(localArtifact.Name())

			Expect(err).NotTo(HaveOccurred())
			Expect(fileContents).To(ContainSubstring("\"my_bucket\": {"))
			Expect(fileContents).To(ContainSubstring("\"bucket_name\": \"" + backupBucket + "\""))
			Expect(fileContents).To(ContainSubstring("\"bucket_region\": \"" + backupRegion + "\""))
			Expect(fileContents).To(MatchRegexp(
				"\"path\": \"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}\\/my_bucket\""))
		})

		DeleteAllFilesFromBucket(region, bucket)
		Expect(ListFilesFromBucket(region, bucket)).To(HaveLen(0))
		WriteFileInBucket(region, bucket, "should/be/left/alone", "STILL_HERE")

		By("restoring from the backup bucket to the source bucket", func() {
			Expect(backuperInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
				" /var/vcap/jobs/s3-unversioned-blobstore-backup-restorer/bin/bbr/restore")).To(Succeed())

			postRestoreFiles = ListFilesFromBucket(region, bucket)
			Expect(postRestoreFiles).To(ConsistOf([]string{"should/be/left/alone", "original/path/to/file"}))
			Expect(GetFileContentsFromBucket(region, bucket, "original/path/to/file")).To(Equal("FILE1"))
			Expect(GetFileContentsFromBucket(region, bucket, "should/be/left/alone")).To(Equal("STILL_HERE"))
		})
	})
})
