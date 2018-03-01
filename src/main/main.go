package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"regexp"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	uuid "github.com/satori/go.uuid"
)

type AccountEvent struct {
	Data      Account `json:"data"`
	EventType string  `json:"eventtype"`
	Source    string  `json:"source"`
	RequestId string  `json:"requestid"`
	// maybe have a version
}

type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

type AccountResponse struct {
	Account Account `json:"account"`
}

type UsersResponse struct {
	Users []User `json:"users"`
}

type UserResponse struct {
	User User `json:"user"`
}

func migrations() migrate.MigrationSource {
	return &migrate.AssetMigrationSource{
		Asset:    Asset,
		AssetDir: AssetDir,
		Dir:      "db/migrations",
	}
}

func static_handler(rw http.ResponseWriter, req *http.Request) {
	var path string = req.URL.Path
	fmt.Println("static_handler")
	if path == "" {
		path = "index.html"
	}
	r, _ := regexp.Compile(".css?")
	if r.Match([]byte(path)) {
		rw.Header().Set("Content-Type", "text/css")
	}
	if bs, err := Asset("static/" + path); err != nil {
		rw.WriteHeader(http.StatusNotFound)
	} else {
		var reader = bytes.NewBuffer(bs)
		io.Copy(rw, reader)
	}
}

//
func indexHandler(client *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("indexHandler")

		var json []byte
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	}
}

//
func usersHandler(client *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("usersHandler")
		log.Println("usersHandler", r.Method)

		res_json := []byte("unknown")
		switch r.Method {
		case "GET":
			var u User
			users, err := u.all(client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			res_json, err = json.Marshal(UsersResponse{Users: users})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "POST":
			decoder := json.NewDecoder(r.Body)
			log.Println("usersHandler", r.Body)

			var u User
			uid := uuid.NewV4().String()
			err := decoder.Decode(&u)
			u.Uid = uid

			u, err = u.create(client)
			res_json, err = json.Marshal(UserResponse{User: u})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Println("usersHandler response", res_json)
		default:
			res_json = []byte("unknown")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res_json)
	}
}

//
func accountsHandler(client *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("accountsHandler")
		log.Println("accountsHandler", r.Method)

		res_json := []byte("unknown")
		switch r.Method {
		case "GET":
			var a Account
			accounts, err := a.all(client)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			res_json, err = json.Marshal(AccountsResponse{Accounts: accounts})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		case "POST":
			decoder := json.NewDecoder(r.Body)
			log.Println("accountsHandler", r.Body)

			var a Account
			uid := uuid.NewV4().String()
			err := decoder.Decode(&a)
			a.Uid = uid
			a, err = a.create(client)
			res_json, err = json.Marshal(AccountResponse{Account: a})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			kafka := newKafkaSyncProducer()
			var ae AccountEvent
			ae.Data = a
			ae.EventType = "account_created"
			ae.Source = "greenscreens"
			ae.RequestId = uid
			sendMsg(kafka, ae)

			log.Println("accountsHandler response", res_json)
		default:
			res_json = []byte("unknown")
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res_json)
	}
}

//
func main() {

	log.Println("connection string: ", os.Getenv("DB_CONNECTION_STRING"))
	client, err := sql.Open("postgres", os.Getenv("DB_CONNECTION_STRING"))
	log.Println("type of client: ", reflect.TypeOf(client))
	if err != nil {
		log.Fatal("cannot connect to database: ", err)
	}
	log.Println("client created")

	log.Println("before ping")
	if err2 := client.Ping(); err2 != nil {
		log.Println("Failed to keep connection alive", err2)
	}
	log.Println("after ping")

	log.Println("before running migrations")
	n, err := migrate.Exec(client, "postgres", migrations(), migrate.Up)
	if err != nil {
		log.Fatal("db migrations failed: ", err)
	}
	log.Println(n, "migrations run")
	//

	log.Println("starting the consumer...")
	go mainConsumer(client)

	log.Println("before root handler")
	log.Println("running on port 8080")
	// http.HandleFunc("/", indexHandler(client))
	http.HandleFunc("/users", usersHandler(client))
	http.HandleFunc("/accounts", accountsHandler(client))
	http.Handle("/", http.StripPrefix("/", http.HandlerFunc(static_handler)))
	http.ListenAndServe(":8080", nil)
	//
}
