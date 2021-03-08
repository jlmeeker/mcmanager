package storage

import (
	"log"
	"os"
	"path/filepath"
)

// AuditWrite writes an entry to the audit log
func AuditWrite(who, what, where string) error {
	var err error
	var dh *os.File

	for err == nil {
		dh, err = os.OpenFile(filepath.Join(STORAGEDIR, "audit.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, DEFAULTFILEPERM)
		defer dh.Close()
		log.SetOutput(dh)
		log.Printf("%s [%s] %s", what, who, where)
		break
	}

	return err
}
