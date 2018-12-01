

wordpress-backuper
====

Wordpress backup tool written Golang.

## Description

wordpress-backuper is CLI tool to backup wordpress with AWS S3. Backup target is mysql dump and wordpress directory.

## Usage

```
$ sudo /usr/local/bin/wp_backup
2018/08/05 13:06:17 Start backup to s3://${BUCKET_NAME}/20180805T130617
2018/08/05 13:06:17 Start dump database
File is saved to /tmp/wp-backup132075555/wordpress.sql
2018/08/05 13:06:17 Start archive wordpress dir
2018/08/05 13:06:59 Start upload backups to S3
2018/08/05 13:09:17 Finish backup to s3://${BUCKET_NAME}/20180805T130617
```

## Install



```
$ cat /etc/wp_backup.env
BACKUP_DIR=/var/www/html
DB_USERNAME=${MYSQL_USERNAME}
DB_PASSWORD=${MYSQL_PASSWORD}
DB_HOSTNAME=${MYSQL_HOSTNAME}
DB_PORT=${MYSQL_PORT}
DB_NAME=${MYSQL_DB_NAME}
AWS_ACCESS_KEY_ID=${ACCESS_KEY_ID}
AWS_SECRET_ACCESS_KEY=${SECRET_ACCESS_KEY}
AWS_REGION=${REGION_NAME}
AWS_BUCKET=${BUCKET_NAME}
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



http://vccw.test/



Build wp_backup for linux:

```
$ env GOOS=linux GOARCH=amd64 go build -o ./vccw/wp_backup
```

## Contribution

## Licence

[MIT](https://github.com/tcnksm/tool/blob/master/LICENCE)

## Author

[takaishi](https://github.com/takaishi)