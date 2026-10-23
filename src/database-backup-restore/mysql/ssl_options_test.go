package mysql_test

import (
	"database-backup-restore/config"
	"database-backup-restore/mysql"
	"strings"
	"testing"
)

func TestDefaultLegacyTLSDoesNotOfferObsoleteCiphers(t *testing.T) {
	options := mysql.NewLegacySSLOptionsProvider(config.TempFolderManager{}).BuildSSLParams(nil)
	if len(options) != 1 || !strings.HasPrefix(options[0], "--ssl-cipher=") {
		t.Fatalf("missing cipher restrictions: %v", options)
	}
	ciphers := strings.Split(strings.TrimPrefix(options[0], "--ssl-cipher="), ":")
	if len(ciphers) == 0 {
		t.Fatal("no compatible ciphers remain")
	}
	for _, cipher := range ciphers {
		if strings.Contains(cipher, "DES") || strings.Contains(cipher, "RC4") {
			t.Fatalf("obsolete cipher offered: %s", cipher)
		}
	}
}
