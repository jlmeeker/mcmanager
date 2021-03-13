package storage

import (
	"log"
	"os"
	"path/filepath"
)

// AuditWrite writes an entry to the audit log
func AuditWrite(who, what, where string) {
	var dh *os.File
	dh, err := os.OpenFile(filepath.Join(STORAGEDIR, "audit.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, DEFAULTFILEPERM)
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.SetOutput(dh)
	log.Printf("%s [%s] %s", what, who, where)
	dh.Close()
}
