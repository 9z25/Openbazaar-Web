package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

//CreateDatabase open db connection
func CreateDatabase() (*sql.DB, error) {

	db, err := sql.Open("mysql", "richie-admin:**OMMITTED**@/test_user")
	if err != nil {
		return nil, err
	}
	return db, nil
}
