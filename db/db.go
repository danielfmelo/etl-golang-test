package db

import (
	"database/sql"
	_ "github.com/lib/pq"

	"app/utils"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/rs/zerolog/log"
)

var db *sql.DB

type Env struct {
	User   string `json:"user"`
	Pass   string `json:"pass"`
	Host   string `json:"host"`
	Port   string `json:"port"`
	DbName string `json:"dbname"`
}

//Return the dabase environment from /env/env.json file
func getEnv() (e Env) {

	file, _ := ioutil.ReadFile("./env/env.json")
	json.Unmarshal(file, &e)
	return e
}

//Create and return the postgres information for the connection
func getDbinfo() string {
	d := getEnv()
	dbinfo := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Pass, d.Host, d.Port, d.DbName)
	return dbinfo
}

//Initialize database connection
func InitDB() {

	dbinfo := getDbinfo()

	var err error
	db, err = sql.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	log.Info().
		Msgf("DATABASE CONNECTION INITIALIZED")
}

//Close database connection
func CloseDB() {
	db.Close()
	log.Info().
		Msgf("DATABASE CONNECTION CLOSED")
}

//Insert new person information
func InsertPerson(p *utils.Worker) error {

	var lastInsertId int

	err := db.QueryRow("INSERT INTO person(name,job,salary,month) VALUES($1,$2,$3,$4) returning id;", p.Name, p.Job, p.Salary, p.Date).Scan(&lastInsertId)

	if err != nil {
		return err
	}
	return nil
}
