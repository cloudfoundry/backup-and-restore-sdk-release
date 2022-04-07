# go-binmock

[![Build Status](https://travis-ci.org/pivotal-cf/go-binmock.svg?branch=master)](https://travis-ci.org/pivotal-cf/go-binmock)

A library to mock interactions with collaborating executables. It works by creating a fake executable that can be orchestrated by the test. The path of the test executable can be injected into the system under test.


### Usage

Creating a mock binary and getting its path:

```golang
mockMonit = binmock.NewBinMock(ginkgo.Fail)
monitPath := mockMonit.Path

thingToTest := NewThingToTest(monitPath)
```

Setting up expected interactions with the binary:

```golang
mockMonit.WhenCalledWith("start", "all").WillExitWith(0)
mockMonit.WhenCalledWith("summary").WillPrintToStdOut(output)
mockMonit.WhenCalledWith("summary").WillPrintToStdOut(output).WillPrintToStdErr("Noooo!").WillExitWith(1)
```

Asserting on the interactions with the binary, after the fact:

```golang
Expect(mockPGDump.Invocations()).To(HaveLen(1))
Expect(mockPGDump.Invocations()[0].Args()).To(Equal([]string{"dbname"}))
Expect(mockPGDump.Invocations()[0].Env()).To(HaveKeyWithValue("PGPASS", "p@ssw0rd"))
```

For a working example you can look at [backup-and-restore-sdk](https://github.com/cloudfoundry-incubator/backup-and-restore-sdk-release/blob/19fa00e61dcdbf15bfe57633798c8f88a8345e9d/src/github.com/cloudfoundry-incubator/database-backup-and-restore/integration_tests/mysql_test.go)
