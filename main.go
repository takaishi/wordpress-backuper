package main

import (
	"database/sql"
	"fmt"
	"github.com/JamesStewy/go-mysqldump"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
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

func main() {
	err := godotenv.Load("/etc/wp_backup.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	n := time.Now()
	dumpSubdir := n.Format("20060102T150405")
	bucket := os.Getenv("AWS_BUCKET")

	log.Printf("Start backup to s3://%s/%s\n", bucket, dumpSubdir)

	dir, err := ioutil.TempDir("", "wp-backup")
	if err != nil {
		log.Fatalf("Failed to create tempdir: %s\n", err)
	}
	defer os.RemoveAll(dir)

	log.Printf("Start dump database\n")
	err = DumpDatabase(dir)
	if err != nil {
		log.Fatalf("Failed to dump database: %s\n", err)
	}

	log.Printf("Start archive wordpress dir\n")
	err = BackupWordpressFiles(dir)
	if err != nil {
		log.Fatalf("Failed to backup wordpress files: %s\n", err)
	}

	log.Printf("Start upload backups to S3\n")
	err = UploadToS3(dir, dumpSubdir)
	if err != nil {
		log.Fatalf("Failed to upload to s3: %s\n", err)
	}

	log.Printf("Finish backup to s3://%s/%s\n", bucket, dumpSubdir)
}

func DumpDatabase(dumpDir string) error {
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	hostname := os.Getenv("DB_HOSTNAME")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, hostname, port, dbname))
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
	fmt.Printf("File is saved to %s\n", resultFilename)

	return nil
}

func BackupWordpressFiles(dumpDir string) error {
	backupDir := os.Getenv("BACKUP_DIR")
	dumpFileFormat := fmt.Sprintf("%s/wordpress.zip", dumpDir)
	err := archiver.Zip.Make(dumpFileFormat, []string{backupDir})
	if err != nil {
		return errors.Wrap(err, "Failed to archive")
	}

	return nil
}

func UploadToS3(dumpDir string, dumpSubdir string) error {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")
	bucket := os.Getenv("AWS_BUCKET")

	paths := []string{"wordpress.zip", "wordpress.sql"}

	for _, f := range paths {
		dumpFileFormat := fmt.Sprintf("%s/%s", dumpDir, f)
		file, err := os.Open(dumpFileFormat)
		if err != nil {
			return errors.Wrap(err, "Failed to open file")
		}
		defer file.Close()
		cli := s3.New(session.New(), &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
			Region:      aws.String(region),
		})

		_, err = cli.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("%s/%s", dumpSubdir, f)),
			Body:   file,
		})

		if err != nil {
			return errors.Wrap(err, "Failed to PutObject")
		}
	}
	return nil
}
