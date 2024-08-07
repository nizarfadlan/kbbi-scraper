package database

import (
	"fmt"
	"os"
	"time"

	"kbbi-scraper/internal/common"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var schema = `
CREATE TABLE IF NOT EXISTS lema (
    id INT AUTO_INCREMENT PRIMARY KEY,
    kata VARCHAR(255) NOT NULL,
    lema VARCHAR(255) NOT NULL,
    kelas_kata TINYTEXT,
    keterangan TEXT
);

CREATE TABLE IF NOT EXISTS words (
    id INT AUTO_INCREMENT PRIMARY KEY,
    kata VARCHAR(255) UNIQUE NOT NULL
);`

type Lema struct {
	Id         int    `db:"id"`
	Kata       string `db:"kata"`
	Lema       string `db:"lema"`
	KelasKata  string `db:"kelas_kata"`
	Keterangan string `db:"keterangan"`
}

type Kata struct {
	Id   int    `db:"id"`
	Kata string `db:"kata"`
}

func ConnectDB() (*sqlx.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local&charset=utf8mb4&multiStatements=true",
		username, password, host, port, dbname)

	db, err := sqlx.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("Failed to ping the database: " + err.Error())
	}

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("error creating schema: %w", err)
	}

	return db, nil
}

func CloseDB(db *sqlx.DB) {
	if err := db.Close(); err != nil {
		common.PrintError("Failed to close database: %v", err)
	}
}

func InsertLemas(db *sqlx.DB, lemas []Lema) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(`
		INSERT INTO lema (kata, lema, kelas_kata, keterangan)
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
	return exists, err
}

func InsertWords(db *sqlx.DB, words []string) error {
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Preparex(`
		INSERT INTO words (kata)
		VALUES (?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, word := range words {
		_, err := stmt.Exec(word)
		if err != nil {
			return fmt.Errorf("failed to insert word %+v: %w", word, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func GetWords(db *sqlx.DB) ([]Kata, error) {
	var words []Kata
	err := db.Select(&words, "SELECT * FROM words")
	return words, err
}
