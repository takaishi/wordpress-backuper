package main

import (
	"database/sql"
	"fmt"
	"github.com/JamesStewy/go-mysqldump"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"log"
)

type Dumper struct {
	db          DBConfig
	wpRootDir   string
	destination string
}

func (d *Dumper) Run() error {
	log.Printf("Start dump database: %s\n", d.destination)
	err := d.dumpDatabase()
	if err != nil {
		log.Fatalf("Failed to dump database: %s\n", err)
		return err
	}

	log.Printf("Start dump wordpress dir\n")
	err = d.backupWordpressFiles()
	if err != nil {
		log.Fatalf("Failed to dump wordpress dir: %s\n", err)
		return err
	}

	return nil
}

func (d *Dumper) dumpDatabase() error {
	username := d.db.Username
	password := d.db.Password
	hostname := d.db.Hostname
	port := d.db.Port
	dbname := d.db.Database

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", username, password, hostname, port, dbname))
	if err != nil {
		return errors.Wrap(err, "Error opening database: %s")
	}

	dumper, err := mysqldump.Register(db, d.destination, "wordpress")
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

func (d *Dumper) backupWordpressFiles() error {
	dumpFileFormat := fmt.Sprintf("%s/wordpress.zip", d.destination)
	err := archiver.Zip.Make(dumpFileFormat, []string{d.wpRootDir})
	if err != nil {
		return errors.Wrap(err, "Failed to archive")
	}

	return nil
}
