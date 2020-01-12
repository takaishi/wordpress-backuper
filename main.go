package main

import (
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
)

var Version string

const DB_DUMP_FILE = "wordpress.sql"
const WORDPRESS_FILE = "wordpress.zip"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	app := cli.NewApp()
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "config.toml",
		},
	}
	app.Action = func(c *cli.Context) error {
		return action(c)
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func action(c *cli.Context) error {
	var config Config
	_, err := toml.DecodeFile(c.String("config"), &config)
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "wp-backup")
	if err != nil {
		log.Fatalf("Failed to create tempdir: %s\n", err)
		return err
	}
	defer os.RemoveAll(dir)

	dumper := Dumper{db: config.DB, wpRootDir: config.Wordpress.RootDir, destination: dir}
	err = dumper.Run()
	if err != nil {
		log.Fatalf("Failed to dump: %s\n", err)
		return err
	}

	if config.AWS != nil {
		backuper := AWSBackuper{source: dir, aws: config.AWS}
		return backuper.Run()
	}

	if config.Local != nil {
		backuper := LocalBackuper{
			source:      dir,
			destination: config.Local.Destination,
		}
		return backuper.Run()
	}
	return nil
}
