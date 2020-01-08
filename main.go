package main

import (
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"log"
	"os"
)

var Version string

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

	if config.AWS != nil {
		backuper := AWSBackuper{cfg: config}
		return backuper.Run()
	}

	if config.Local != nil {
		backuper := LocalBackuper{
			db:          config.DB,
			wpRootDir:   config.Wordpress.RootDir,
			destination: config.Local.Destination,
		}
		return backuper.Run()
	}
	return nil
}
