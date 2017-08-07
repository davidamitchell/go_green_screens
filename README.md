# go_green_screens


## Database migrations


### Preconditions
export DB_CONNECTION_STRING="dbname=greenscreen_dev user=765440 sslmode=disable"
go-bindata - go get -u github.com/jteeuwen/go-bindata/...
go build gb https://getgb.io/docs/install/


gb vendor fetch github.com/rubenv/sql-migrate
gb vendor fetch gopkg.in/gorp.v1






* have go-bindata installed
* `gb vendor fetch github.com/rubenv/sql-migrate`
* `create database docker_test;`
* `create database docker_test_test;`
* `create database docker_test_developement;`
* `go-bindata -pkg main -o bindata.go db/migrations/`

We are using [goose](https://github.com/ox/goose) to manage db migrations (note, this is not to be confused with https://bitbucket.org/liamstask/goose).

To install:
```
go get github.com/ox/goose/cmd/goose
```
This will add a `goose` executable to your `$GOPATH/bin` directory.

To create the database as defined in the `db/dbconf.yml` file (the `development` environment is the default, to change use the `-env="environment"` switch):
```
 goose create-db
```

To then run outstanding migrations:
```
 goose up
```
