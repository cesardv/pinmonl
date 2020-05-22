package database

import (
	"database/sql"
	"sync"

	"github.com/jmoiron/sqlx"
)

// DB stores sql.DB and the driver name.
type DB struct {
	*sqlx.DB
	Locker
}

// Open creates db.
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, driverName)

	var locker Locker
	switch driverName {
	case "sqlite3":
		locker = &sync.RWMutex{}
	default:
		locker = &nopLocker{}
	}

	return &DB{dbx, locker}, nil
}

// Beginx creates database transaction.
func (db *DB) Beginx() (*Tx, error) {
	db.Lock()
	tx, err := db.DB.Beginx()
	if err != nil {
		db.Unlock()
		return nil, err
	}
	return &Tx{tx, db.Locker}, nil
}

// NamedQuery wraps database's query.
func (db *DB) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	db.Lock()
	rows, err := db.DB.NamedQuery(query, arg)
	db.Unlock()
	return rows, err
}

// Locker represents an object that can be locked and unlocked.
type Locker interface {
	Lock()
	RLock()
	RUnlock()
	Unlock()
}

type nopLocker struct{}

func (nopLocker) Lock()    {}
func (nopLocker) RLock()   {}
func (nopLocker) RUnlock() {}
func (nopLocker) Unlock()  {}

// Tx handles database transaction.
type Tx struct {
	*sqlx.Tx
	Locker
}

// NamedQuery wraps transaction's query.
func (tx *Tx) NamedQuery(query string, arg interface{}) (*sqlx.Rows, error) {
	rows, err := tx.Tx.NamedQuery(query, arg)
	return rows, err
}

// Commit commits database transaction.
func (tx *Tx) Commit() error {
	err := tx.Tx.Commit()
	tx.Unlock()
	return err
}

// Rollback rollbacks database transaction.
func (tx *Tx) Rollback() error {
	err := tx.Tx.Rollback()
	tx.Unlock()
	return err
}

// Binder provides database bind var functions.
type Binder interface {
	Rebind(string) string
	BindNamed(string, interface{}) (string, []interface{}, error)
}

// Execer extends sqlx.Execer.
type Execer interface {
	sqlx.Execer
	Binder
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

// Queryer extends sqlx.Queryer.
type Queryer interface {
	sqlx.Queryer
	Binder
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}

// Ext combines Execer and Queryer.
type Ext interface {
	sqlx.Execer
	sqlx.Queryer
	Binder
	DriverName() string
	NamedExec(query string, arg interface{}) (sql.Result, error)
	NamedQuery(query string, arg interface{}) (*sqlx.Rows, error)
}
