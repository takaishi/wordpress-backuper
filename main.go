package main

import (
	"database/sql"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
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

	n := time.Now()
	dumpSubdir := n.Format("20060102T150405")
	bucket := config.AWS.Bucket

	log.Printf("Start backup to s3://%s/%s\n", bucket, dumpSubdir)

	dir, err := ioutil.TempDir("", "wp-backup")
	if err != nil {
		log.Fatalf("Failed to create tempdir: %s\n", err)
		return err
	}
	defer os.RemoveAll(dir)

	log.Printf("Start dump database: %s\n", dir)
	err = DumpDatabase(config, dir)
	if err != nil {
		log.Fatalf("Failed to dump database: %s\n", err)
		return err
	}

	log.Printf("Start archive wordpress dir\n")
	err = BackupWordpressFiles(config, dir)
	if err != nil {
		log.Fatalf("Failed to backup wordpress files: %s\n", err)
		return err
	}

	log.Printf("Start upload backups to S3\n")
	err = BackupToS3(config, dir, dumpSubdir)
	if err != nil {
		log.Fatalf("Failed to upload to s3: %s\n", err)
		return err
	}

	log.Printf("Start rotate backups\n")
	err = RotateBackup(config)
	if err != nil {
		log.Fatalf("Failed to rotate backups: %s\n", err)
		return err
	}

	log.Printf("Finish backup to s3://%s/%s\n", bucket, dumpSubdir)

	return nil
}

func DumpDatabase(config Config, dumpDir string) error {
	username := config.DB.Username
	password := config.DB.Password
	hostname := config.DB.Hostname
	port := config.DB.Port
	dbname := config.DB.Database

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

func BackupWordpressFiles(config Config, dumpDir string) error {
	backupDir := config.Wordpress.RootDir
	dumpFileFormat := fmt.Sprintf("%s/wordpress.zip", dumpDir)
	err := archiver.Zip.Make(dumpFileFormat, []string{backupDir})
	if err != nil {
		return errors.Wrap(err, "Failed to archive")
	}

	return nil
}

func BackupToS3(config Config, dumpDir string, dumpSubdir string) error {
	accessKeyID := config.AWS.AccessKeyID
	secretAccessKey := config.AWS.SecretAccessKey
	region := config.AWS.Region
	bucket := config.AWS.Bucket

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to create session")
	}
	cli := s3.New(sess)

	for _, name := range []string{"wordpress.sql", "wordpress.zip"} {
		err := UploadToS3(cli, fmt.Sprintf("%s/%s", dumpDir, name), bucket, fmt.Sprintf("%s/%s", dumpSubdir, name))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to upload %s to s3", name))
		}
	}
	return nil

}

func UploadToS3(cli *s3.S3, path string, bucket string, key string) error {
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "Failed to open file")
	}
	defer f.Close()

	_, err = cli.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   f,
	})

	if err != nil {
		return errors.Wrap(err, "Failed to PutObject")
	}

	return nil
}

func RotateBackup(config Config) error {
	backup_size := 3
	accessKeyID := config.AWS.AccessKeyID
	secretAccessKey := config.AWS.SecretAccessKey
	region := config.AWS.Region
	bucket := config.AWS.Bucket

	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Region:      aws.String(region),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to create session")
	}

	cli := s3.New(sess)

	keys, err := GetDeletePrefixes(cli, bucket, backup_size)
	if err != nil {
		return errors.Wrap(err, "Failed to get delete prefixes")
	}

	for _, k := range keys {
		result, err := cli.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:    aws.String(bucket),
			Prefix:    aws.String(fmt.Sprintf("%s", k)),
			Delimiter: aws.String("/"),
		})
		if err != nil {
			return errors.Wrap(err, "Failed to list object")
		}
		for _, o := range result.Contents {
			err := DeleteObject(cli, bucket, *o.Key)
			if err != nil {
				return errors.Wrap(err, "Failed to DeleteObject")
			}
		}
	}
	return nil
}

func GetDeletePrefixes(cli *s3.S3, bucket string, backup_size int) ([]string, error) {
	var keys []string
	result, err := cli.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	})

	if err != nil {
		return nil, errors.Wrap(err, "Failed to list object")
	}

	for _, k := range result.CommonPrefixes[0 : len(result.CommonPrefixes)-backup_size] {
		keys = append(keys, *k.Prefix)
	}

	return keys, nil
}

func DeleteObject(cli *s3.S3, bucket string, key string) error {
	log.Printf("Delete Object: s3://%s/%s\n", bucket, key)

	_, err := cli.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return errors.Wrap(err, "Failed to delete object")
	}
	return nil
}
