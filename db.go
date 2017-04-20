package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const createSQLlitestmt = `
PRAGMA foreign_keys = off;
BEGIN TRANSACTION;

-- Table: Projects
CREATE TABLE Projects (projectid INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL, name TEXT NOT NULL, explanation TEXT NOT NULL, goal DOUBLE NOT NULL);

-- Table: Donations
CREATE TABLE Donations (projectid INTEGER REFERENCES Projects (projectid) ON DELETE CASCADE ON UPDATE CASCADE NOT NULL, name TEXT NOT NULL, amount DOUBLE NOT NULL, message TEXT NOT NULL);

COMMIT TRANSACTION;
PRAGMA foreign_keys = on;

`

var db *sqlx.DB

type dbConnection struct {
	Driver     string
	Connection string
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func initialisation(dbc dbConnection) {
	var err error
	db, err = sqlx.Open(dbc.Driver, dbc.Connection)
	if err != nil {
		log.Fatal(err)
	}
	initDB(dbc)
}

func initDB(dbc dbConnection) {
	switch dbc.Driver {
	case "sqlite3":
		cont, err := exists(dbc.Connection)
		if err != nil {
			log.Fatal(err)
		}
		if cont {
			fmt.Println("cont")
			return
		}
		_, err = os.Create(dbc.Connection)
		if err != nil {
			log.Fatal("Could not create file "+dbc.Connection, err)
		}
		_, err = db.Exec(createSQLlitestmt)
		if err != nil {
			log.Printf("%q: %s\n", err, createSQLlitestmt)
			return
		}
	default:
		log.Fatal("DB Driver unkown. Stopping Server")
	}
}

func GetProjects() ([]project, error) {
	var res []project
	err := db.Select(&res, "Select * from Projects")
	if err != nil {
		return res, errors.New("Error getting Projects:" + err.Error())
	}
	for i := range res {
		err := db.Select(&res[i].Donations, "Select Amount, Message, Name from Donations Where projectID=?", res[i].ProjectID)
		if err != nil {
			return res, err
		}
	}
	return res, nil
}
