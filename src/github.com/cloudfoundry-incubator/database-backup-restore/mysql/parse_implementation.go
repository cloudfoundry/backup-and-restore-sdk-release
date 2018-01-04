package mysql

import "fmt"

func ParseImplementation(versionComment string) (string, error) {
	switch versionComment {
	case "MySQL Community Server (GPL)":
		return "mysql", nil
	case "MariaDB Server":
		return "mariadb", nil
	default:
		return "", fmt.Errorf("unsupported mysql database: %s", versionComment)
	}
}
