package main

type Config struct {
	DB        DBConfig        `toml:"DB"`
	Wordpress WordpressConfig `toml:"Wordpress"`
	AWS       *AWSConfig      `toml:"AWS"`
	Local     *LocalConfig    `toml:"Local"`
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

type LocalConfig struct {
	Destination string `toml:"destination"`
}
