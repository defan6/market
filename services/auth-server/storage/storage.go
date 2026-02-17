package db

import (
	"github.com/jmoiron/sqlx"
)

type Database struct {
	db *sqlx.DB
}

func NewDatabase() *Database {
	db, err := sqlx.Open("postgres", "postgres://postgres:postgres@localhost:5433/users?sslmode=disable")
	if err != nil {
		panic(err)
	}
	return &Database{db: db}
}

func (d *Database) GetDB() *sqlx.DB {
	return d.db
}
