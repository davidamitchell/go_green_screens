package main

import (
	"database/sql"
	"log"
	"time"
)

type User struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name"`
	Created time.Time `json:"created"`
	Uid     string    `json:"uid"`
}

type Account struct {
	Id      int64     `json:"id"`
	Name    string    `json:"name"`
	Owner   string    `json:"owner"`
	Created time.Time `json:"created"`
	Uid     string    `json:"uid"`
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (u User) all(c *sql.DB) ([]User, error) {

	var users = make([]User, 0)

	rows, err := c.Query(`select id, name, created from users order by created;`)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var user User
		err = rows.Scan(&user.Id, &user.Name, &user.Created)
		checkErr(err)
		users = append(users, user)
	}
	err = rows.Err()
	checkErr(err)
	return users, nil

}

///////////////////////////
func (u User) find(c *sql.DB) (User, error) {

	var user User

	rows, err := c.Query(`
		select id, name, created, uid
		from users
		where name = $1;
		`, u.Name)
	checkErr(err)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&user.Id, &user.Name, &user.Created, &user.Uid)
		checkErr(err)
	}
	err = rows.Err()
	checkErr(err)

	return user, nil
}

func (u User) create(c *sql.DB) (User, error) {
	log.Println("user create", u.Name)
	var created = time.Now().Format(time.RFC3339)

	// stmt, err := c.Prepare(`insert into users (name, created) values ($1, $2) RETURNING id;`)
	// log.Println("stam", stmt)
	// checkErr(err)
	var id int64
	err := c.QueryRow(`
								insert into users
								(name, created, uid)
								values
								($1, $2, $3)
								RETURNING id;
								`, u.Name, created, u.Uid).Scan(&id)
	checkErr(err)

	u.Id = id
	u.Created, err = time.Parse(time.RFC3339, created)
	checkErr(err)
	return u, nil
}

///////////////////////////
///////////////////////////
// Accounts
///////////////////////////
///////////////////////////
func (u Account) all(c *sql.DB) ([]Account, error) {

	var accounts = make([]Account, 0)

	rows, err := c.Query(`
		select a.id, a.name, u.name as owner, a.created, a.uid
		from accounts a join users u on u.id = a.users_id
		order by a.created;
		`)
	checkErr(err)
	defer rows.Close()

	for rows.Next() {
		var a Account
		err = rows.Scan(&a.Id, &a.Name, &a.Owner, &a.Created, &a.Uid)
		checkErr(err)
		accounts = append(accounts, a)
	}
	err = rows.Err()
	checkErr(err)
	return accounts, nil

}

///////////////////////////
func (a Account) find(c *sql.DB) (Account, error) {

	var account Account

	rows, err := c.Query(`
		select a.id, a.name, u.name as owner, a.created, a.uid
		from accounts a join users u on u.id = a.users_id
		and u.name = $2
		 	where a.name = $1;
		`, a.Name, a.Owner)
	checkErr(err)

	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&account.Id, &account.Name, &account.Owner, &account.Created, &account.Uid)
		checkErr(err)
	}
	err = rows.Err()
	checkErr(err)

	return account, nil
}

func (a Account) create(c *sql.DB) (Account, error) {
	log.Println("account create", a.Name)
	var created = time.Now().Format(time.RFC3339)

	account, err := a.find(c)

	var u User
	u.Name = a.Owner
	u.Uid = a.Uid
	user, err := u.find(c)

	if user.Id <= 0 {
		u, err = u.create(c)
		checkErr(err)
	}

	if account.Id <= 0 {
		var id int64
		err = c.QueryRow(`
				insert into accounts (name, users_id, created, uid)
				select $1, id, $3, $4
				from users
				where name = $2
				RETURNING id;
				`, a.Name, a.Owner, created, a.Uid).Scan(&id)
		checkErr(err)

		a.Id = id
		a.Created, err = time.Parse(time.RFC3339, created)
		checkErr(err)
	} else {
		a = account
	}
	return a, nil
}
