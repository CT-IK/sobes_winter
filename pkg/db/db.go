// Package db utilizes sqlx library for convenient work with databases and their global state
package db

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"log"
)

// Mode is simple string that represent mode of Database to initialize
type Mode string

const (
	ProdMode Mode = "prod"
	DevMode  Mode = "dev"
	TestMode Mode = "test"
)

// Config represents configuration of database to initialize
type Config struct {
	Mode       Mode
	DriverName string
	DSN        string
}

var db *sqlx.DB

// NewProdConfig returns new db.Config for database with db.ProdMode, that means any real db that you want
func NewProdConfig(driverName string, dsn string) Config {
	return Config{
		Mode:       ProdMode,
		DriverName: driverName,
		DSN:        dsn,
	}
}

// NewDevConfig returns new db.Config for database with db.DevMode, that means in-memory SQLite database
func NewDevConfig() Config {
	return Config{
		Mode:       DevMode,
		DriverName: "sqlite",
		DSN:        ":memory:",
	}
}

// NewTestConfig returns new db.Config for database with db.TestMode, that means empty database from sqlmock
func NewTestConfig() Config {
	return Config{Mode: TestMode}
}

// Get returns global sqlx.DB or panics if it wasn't initialized.
func Get() *sqlx.DB {
	if db == nil {
		log.Panicln("db is not initialized")
	}

	return db
}

// Initialize accepts DSN string and creates new sqlx.DB which is stored as global.
// Returns created sqlx.DB or error if something went wrong.
func Initialize(config Config) (*sqlx.DB, error) {
	if db != nil {
		return nil, errors.New("db already initialized")
	}

	var (
		base *sqlx.DB
		err  error
	)

	switch config.Mode {
	case ProdMode:
		base, err = sqlx.Connect(config.DriverName, config.DSN)
	case DevMode:
		base, err = sqlx.Connect(config.DriverName, config.DSN)
	case TestMode:
		mock, _, mockErr := sqlmock.New()
		err = mockErr
		base = sqlx.NewDb(mock, "sqlmock")
	default:
		return nil, errors.New("invalid mode")
	}

	if err != nil {
		return nil, err
	}

	db = base
	return db, nil
}

// Close closes and clears global sqlx.DB or panics if it wasn't initialized.
func Close() {
	if db == nil {
		log.Panicln("db is not initialized")
	}

	db.Close() //nolint:errcheck
	db = nil
}
