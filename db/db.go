package db

import (
	"database/sql"
	_ "github.com/lib/pq"

	"app/utils"
	"fmt"
	"io/ioutil"
	"encoding/json"
)

var db *sql.DB

type Env struct {
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	DbName string `json:"dbname"`
}

func getEnv() (e Env) {

	file, _ := ioutil.ReadFile("./env/env.json")
	json.Unmarshal(file, &e)
	return e
}

func getDbinfo() string {
	d := getEnv()

	dbinfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Pass, d.Host, d.Port, d.DbName)
	fmt.Println(dbinfo)
	return dbinfo
}

func InitDB() {

	dbinfo := getDbinfo()

	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func CloseDB() {
	db.Close()
}

func InsertPerson(p *utils.Worker) error {

	var lastInsertId int

	err := db.QueryRow("INSERT INTO person(name,job,salary,month) VALUES($1,$2,$3,$4) returning id;", p.Name, p.Job, p.Salary, p.Date).Scan(&lastInsertId)

	if err != nil {
		return err
	}
	fmt.Printf("Inserted person:  %s 				with ID:	%d\n", p.Name, lastInsertId)
	return nil
}