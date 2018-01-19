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

func ConnectMysql(dbHostname, dbPassword, dbUsername, dbPort, proxyHost, proxyUsername, proxyPrivateKey string) (*sql.DB, *gexec.Session) {
	var proxySession *gexec.Session
	var err error

	var hostname, port = dbHostname, dbPort
	if proxyHost != "" {
		hostname, port = "127.0.0.1", "13306"
		proxySession, err = startTunnel(port, dbHostname, dbPort, proxyUsername, proxyHost, proxyPrivateKey)
		Expect(err).NotTo(HaveOccurred())
	}

	connection, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/", dbUsername, dbPassword, hostname, port))
	Expect(err).NotTo(HaveOccurred())

	return connection, proxySession
}

func ConnectPostgres(dbHostname, dbPassword, dbUsername, dbPort, dbName, proxyHost, proxyUsername, proxyPrivateKey string) (*sql.DB, *gexec.Session) {
	var proxySession *gexec.Session
	var err error

	var hostname, port = dbHostname, dbPort
	if proxyHost != "" {
		hostname, port = "127.0.0.1", "13306"
		proxySession, err = startTunnel(port, dbHostname, dbPort, proxyUsername, proxyHost, proxyPrivateKey)
		Expect(err).NotTo(HaveOccurred())
	}

	connection, err := sql.Open("postgres", fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable dbname=%s", dbUsername, dbPassword, hostname, port, dbName))
	Expect(err).NotTo(HaveOccurred())

	return connection, proxySession
}

func startTunnel(localPort string, remoteHost string, remotePort string, proxyUsername string, proxyHost string, proxyPrivateKey string) (*gexec.Session, error) {
	var err error
	proxySession, err := gexec.Start(exec.Command(
		"ssh",
		"-L",
		fmt.Sprintf("%s:%s:%s", localPort, remoteHost, remotePort),
		proxyUsername+"@"+proxyHost,
		"-i", proxyPrivateKey,
		"-N",
		"-o",
		"UserKnownHostsFile=/dev/null",
		"-o",
		"StrictHostKeyChecking=no",
	), ginkgo.GinkgoWriter, ginkgo.GinkgoWriter)

	time.Sleep(1 * time.Second)

	return proxySession, err
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
