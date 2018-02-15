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

func ConnectMysql(dbHostname string, dbPassword string, dbUsername string, dbPort int, proxyHost string, proxyUsername string, proxyPrivateKey string) (*sql.DB, *gexec.Session) {
	var proxySession *gexec.Session
	var err error

	var hostname, port = dbHostname, dbPort
	if proxyHost != "" {
		hostname, port = "127.0.0.1", 13306
		proxySession, err = startTunnel(port, dbHostname, dbPort, proxyUsername, proxyHost, proxyPrivateKey)
		Expect(err).NotTo(HaveOccurred())
	}

	connection, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbUsername, dbPassword, hostname, port))
	Expect(err).NotTo(HaveOccurred())

	return connection, proxySession
}

type PostgresConnection struct {
	hostname string
	port     int
	username string
	password string
	db       *sql.DB

	proxyHost       string
	proxyUsername   string
	proxyPrivateKey string
	proxySession    *gexec.Session
}

func NewPostgresConnection(hostname string, port int, username, password, proxyHost, proxyUsername, proxyPrivateKey string) *PostgresConnection {
	return &PostgresConnection{
		hostname:        hostname,
		port:            port,
		username:        username,
		password:        password,
		proxyHost:       proxyHost,
		proxyUsername:   proxyUsername,
		proxyPrivateKey: proxyPrivateKey,
	}
}

func (c *PostgresConnection) OpenSuccessfully(dbName string) {
	err := c.Open(dbName)
	Expect(err).NotTo(HaveOccurred())
}

func (c *PostgresConnection) Open(dbName string) error {
	var db *sql.DB
	var proxySession *gexec.Session
	var err error

	hostname := c.hostname
	port := c.port

	if c.proxyHost != "" {
		hostname, port = "127.0.0.1", 13306
		proxySession, err = startTunnel(port, c.hostname, c.port, c.proxyUsername, c.proxyHost, c.proxyPrivateKey)
		Expect(err).NotTo(HaveOccurred())
	}

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s", c.username, c.password, hostname, port, dbName)

	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	_, err = db.Exec("SELECT VERSION();")
	if err != nil {
		return err
	}

	c.db = db
	c.proxySession = proxySession

	return nil
}

func (c *PostgresConnection) RunSQLCommand(command string) {
	_, err := c.db.Exec(command)
	Expect(err).NotTo(HaveOccurred())
}

func (c *PostgresConnection) FetchSQLColumn(command string) []string {
	var returnValue []string
	rows, err := c.db.Query(command)
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

func (c *PostgresConnection) Close() {
	err := c.db.Close()
	Expect(err).NotTo(HaveOccurred())

	if c.proxySession != nil {
		c.proxySession.Kill()
	}
}

func (c *PostgresConnection) SwitchToDb(dbName string) {
	err := c.db.Close()
	Expect(err).NotTo(HaveOccurred())

	hostname := c.hostname
	port := c.port
	if c.proxyHost != "" {
		hostname, port = "127.0.0.1", 13306
	}

	connectionString := fmt.Sprintf("user=%s password=%s host=%s port=%d sslmode=disable dbname=%s", c.username, c.password, hostname, port, dbName)

	db, err := sql.Open("postgres", connectionString)
	Expect(err).NotTo(HaveOccurred())

	c.db = db
}

func startTunnel(localPort int, remoteHost string, remotePort int, proxyUsername string, proxyHost string, proxyPrivateKey string) (*gexec.Session, error) {
	var err error
	proxySession, err := gexec.Start(exec.Command(
		"ssh",
		"-L",
		fmt.Sprintf("%d:%s:%d", localPort, remoteHost, remotePort),
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
