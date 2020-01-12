

wordpress-backuper
====

Wordpress backup tool written Golang.

## Description

wordpress-backuper is CLI tool to backup wordpress with AWS S3. Backup target is mysql dump and wordpress directory.

## Usage

```
$ wordpress-backuper --config config.toml
2018/08/05 13:06:17 Start backup to s3://${BUCKET_NAME}/20180805T130617
2018/08/05 13:06:17 Start dump database
File is saved to /tmp/wp-backup132075555/wordpress.sql
2018/08/05 13:06:17 Start archive wordpress dir
2018/08/05 13:06:59 Start upload backups to S3
2018/08/05 13:09:17 Finish backup to s3://${BUCKET_NAME}/20180805T130617
```

## Install

```
$ cat config.toml
[DB]
username = "MYSQL_USERNAME"
password = "MYSQL_PASSWORD"
hostname = "MYSQL_HOSTNAME"
port     = MYSQL_PORT
database = "MYSQL_DB_NAME"

[Wordpress]
root_dir = "/var/www/html"

[Local]
destination = "/tmp/backup"

[AWS]
access_key_id = "ACCESS_KEY_ID"
secret_access_key = "SECRET_ACCESS_KEY"
region = "REGION_NAME"
bucket = "BUCKET_NAME"
BACKUP_DIR = "/var/www/html"
```

## Development

Sample development environment with [vccw](http://vccw.cc/):

```
$ vagrant plugin install vagrant-hostsupdater
$ vagrant box add vccw-team/xenial64
$ wget https://github.com/vccw-team/vccw/releases/download/3.18.0/vccw-3.18.0.zip
$ unzip vccw-3.18.0.zip
$ vagrant up
```

Build wordpress-backuper for linux:

```
$ env GOOS=linux GOARCH=amd64 go build -o ./vccw/wordpress-backuper
```

## Licence

[MIT](https://github.com/tcnksm/tool/blob/master/LICENCE)

## Author

[takaishi](https://github.com/takaishi)