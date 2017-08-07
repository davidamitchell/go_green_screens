DB_CONNECTION_STRING ?= "dbname=greenscreen_dev user=765440 sslmode=disable"

.DEFAULT_GOAL := run

run: build
	bin/main

build: clean bindata
	gb build main/...

clean:
	if [ -a src/main/bindata.go ]; then rm -f src/main/bindata.go; fi;
	if [ -a bin ]; then rm -rf bin; fi;
	if [ -a pkg ]; then rm -rf pkg; fi;

bindata:
	go-bindata -o src/main/bindata.go db/migrations/

dbcreate:
