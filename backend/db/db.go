package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	connStr := "postgresql://mehedi:1670@localhost:5432/mehedi?sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	DB.SetMaxIdleConns(5)
	DB.SetMaxOpenConns(10)

	createTables()
}

func createTables() {
	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
	)
	`

	_, err := DB.Exec(createUsersTable)
	if err != nil {
		panic("error creating users table")
	}

	createEventsTable := `
	CREATE TABLE IF NOT EXISTS events (
	    id SERIAL PRIMARY KEY,
	    name TEXT NOT NULL,
	    description TEXT NOT NULL,
	    location TEXT NOT NULL,
	    dateTime TIMESTAMP NOT NULL,
	    user_id INTEGER,
	    FOREIGN KEY(user_id) REFERENCES users(id)
	)`

	_, err = DB.Exec(createEventsTable)
	if err != nil {
		panic("error creating events table")
	}

	createRegistrationsTable := `
	CREATE TABLE IF NOT EXISTS registrations (
	    id SERIAL PRIMARY KEY,
	    event_id INTEGER NOT NULL,
	    user_id INTEGER,
	    FOREIGN KEY(user_id) REFERENCES users(id),
	    FOREIGN KEY(event_id) REFERENCES events(id)
	)
	`

	_, err = DB.Exec(createRegistrationsTable)
	if err != nil {
		panic("error creating registration table")
	}
}
