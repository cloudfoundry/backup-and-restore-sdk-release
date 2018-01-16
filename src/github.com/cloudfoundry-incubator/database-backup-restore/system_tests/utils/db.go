package utils

import (
	"database/sql"
	"fmt"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func Connect(dbDriverProvider DBDriverProvider, dbHostname, dbPassword, dbUsername, dbPort, proxyHost, proxyUsername, proxyPrivateKey string) (*sql.DB, *gexec.Session) {

	if proxyHost != "" {
		proxiedDBHostName := "127.0.0.1"
		proxiedDBPort := "13306"
		var err error
		proxySession, err := gexec.Start(exec.Command(
			"ssh",
			"-L",
			fmt.Sprintf("%s:%s:%s", proxiedDBPort, dbHostname, dbPort),
			proxyUsername+"@"+proxyHost,
			"-i", proxyPrivateKey,
			"-N",
			"-o",
			"UserKnownHostsFile=/dev/null",
			"-o",
			"StrictHostKeyChecking=no",
		), ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)

		connection, err := sql.Open(dbDriverProvider(dbUsername, dbPassword, proxiedDBHostName, proxiedDBPort))
		Expect(err).NotTo(HaveOccurred())
		return connection, proxySession
	} else {
		connection, err := sql.Open(dbDriverProvider(dbUsername, dbPassword, dbHostname, dbPort))

		Expect(err).NotTo(HaveOccurred())
		return connection, nil
	}
}

func ConnectPostgres(dbHostname, dbPassword, dbUsername, dbPort, dbName, proxyHost, proxyUsername, proxyPrivateKey string) (*sql.DB, *gexec.Session) {
	if proxyHost != "" {
		proxiedDBHostName := "127.0.0.1"
		proxiedDBPort := "13306"
		var err error
		proxySession, err := gexec.Start(exec.Command(
			"ssh",
			"-L",
			fmt.Sprintf("%s:%s:%s", proxiedDBPort, dbHostname, dbPort),
			proxyUsername+"@"+proxyHost,
			"-i", proxyPrivateKey,
			"-N",
			"-o",
			"UserKnownHostsFile=/dev/null",
			"-o",
			"StrictHostKeyChecking=no",
		), ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		time.Sleep(1 * time.Second)

		connection, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s", dbUsername, dbPassword, proxiedDBHostName, proxiedDBPort, dbName))
		Expect(err).NotTo(HaveOccurred())
		return connection, proxySession
	} else {
		connection, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s", dbUsername, dbPassword, dbHostname, dbPort, dbName))

		fmt.Println("Connecting database " + dbName)
		Expect(err).NotTo(HaveOccurred())
		return connection, nil
	}
}

type DBDriverProvider func(dbUsername, dbPassword, dbHostname, dbPort string) (string, string)

func MySQL(dbUsername, dbPassword, dbHostname, dbPort string) (string, string) {
	return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUsername, dbPassword, dbHostname, dbPort)
}
func Postgres(dbUsername, dbPassword, dbHostname, dbPort string) (string, string) {
	return "postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=template1", dbUsername, dbPassword, dbHostname, dbPort)
}

func RunSQLCommand(command string, connection *sql.DB) {
	_, err := connection.Exec(command)
	Expect(err).NotTo(HaveOccurred())
}

func FetchSQLColumn(command string, connection *sql.DB) []string {
	var returnValue []string
	rows, err := connection.Query(command)
	Expect(err).NotTo(HaveOccurred())

	defer rows.Close()
	for rows.Next() {
		var rowData string
		Expect(rows.Scan(&rowData)).NotTo(HaveOccurred())

		returnValue = append(returnValue, rowData)
	}
	Expect(rows.Err()).NotTo(HaveOccurred())
	return returnValue
}
