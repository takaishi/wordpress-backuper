package main

import (
	"database/sql"
	"fmt"
	"github.com/JamesStewy/go-mysqldump"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"log"
)

type LocalBackuper struct {
	db          DBConfig
	wpRootDir   string
	destination string
}

func (b *LocalBackuper) Run() error {
	log.Printf("Start dump database: %s\n", b.destination)
	err := b.dumpDatabase(b.destination)
	if err != nil {
		log.Fatalf("Failed to dump database: %s\n", err)
		return err
	}

	log.Printf("Start archive wordpress dir\n")
	err = b.backupWordpressFiles(b.destination)
	if err != nil {
		log.Fatalf("Failed to backup wordpress files: %s\n", err)
		return err
	}

	return nil
}

func (b *LocalBackuper) dumpDatabase(dumpDir string) error {
	username := b.db.Username
	password := b.db.Password
	hostname := b.db.Hostname
	port := b.db.Port
	dbname := b.db.Database

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, hostname, port, dbname))
	if err != nil {
		return errors.Wrap(err, "Error opening database: %s")
	}

	dumper, err := mysqldump.Register(db, dumpDir, "wordpress")
	if err != nil {
		return errors.Wrap(err, "Error registering databse: %s")
	}
	defer dumper.Close()

	resultFilename, err := dumper.Dump()
	if err != nil {
		return errors.Wrap(err, "Error dumping: %s")
	}
	log.Printf("File is saved to %s\n", resultFilename)

	return nil
}

func (b *LocalBackuper) backupWordpressFiles(dumpDir string) error {
	dumpFileFormat := fmt.Sprintf("%s/wordpress.zip", dumpDir)
	err := archiver.Zip.Make(dumpFileFormat, []string{b.wpRootDir})
	if err != nil {
		return errors.Wrap(err, "Failed to archive")
	}

	return nil
}
