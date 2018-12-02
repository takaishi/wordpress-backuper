package main

import (
	"github.com/BurntSushi/toml"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"log"
	"os"
)

type Config struct {
	DB        DBConfig        `toml:"DB"`
	Wordpress WordpressConfig `toml:"Wordpress"`
	AWS       AWSConfig       `toml:"AWS"`
}

type DBConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
	Hostname string `toml:"hostname"`
	Port     int    `toml:"port"`
	Database string `toml:"database"`
}

type WordpressConfig struct {
	RootDir string `toml:"root_dir"`
}

type AWSConfig struct {
	AccessKeyID     string `toml:"access_key_id"`
	SecretAccessKey string `toml:"secret_access_key"`
	Region          string `toml:"region"`
	Bucket          string `toml:"bucket"`
}

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

	backuper := Backuper{cfg: config}

	return backuper.Run()
}
