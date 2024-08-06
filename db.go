package main

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Lema struct {
	Id         int    `db:"id"`
	Kata       string `db:"kata"`
	Lema       string `db:"lema"`
	KelasKata  string `db:"kelas_kata"`
	Keterangan string `db:"keterangan"`
}

func ConnectDB() (*sqlx.DB, error) {
	dsn := "root:nizar@(localhost:3306)/kbbi"

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping the database: " + err.Error())
	}

	return db, nil
}

func InsertLemas(db *sqlx.DB, lemas []Lema) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(`
		INSERT INTO lema (kata, lema, kelas_kata, arti_kata)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, lema := range lemas {
		_, err := stmt.Exec(lema.Kata, lema.Lema, lema.KelasKata, lema.Keterangan)
		if err != nil {
			return fmt.Errorf("failed to insert lema %+v: %w", lema, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func ExistsLemaByKata(db *sqlx.DB, kata string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM lema WHERE kata = ?)`
	var exists bool
	err := db.Get(&exists, query, kata)
	if err != nil {
		return false, err
	}

	return exists, nil
}
