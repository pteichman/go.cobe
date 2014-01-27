package cobe

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	// Blank import loads go-sqlite3 support into database/sql.
	_ "github.com/mattn/go-sqlite3"
)

type graphOptions struct {
	Order     int
	tokenizer string
}

var defaultGraphOptions = &graphOptions{3, "Cobe"}

func InitGraph(path string, opts *graphOptions) error {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	defer db.Close()

	log.Println("Creating table: info")
	_, err = db.Exec(`
CREATE TABLE info (
	attribute TEXT NOT NULL PRIMARY KEY,
	text TEXT NOT NULL)`)

	if err != nil {
		return err
	}

	log.Println("Creating table: tokens")
	_, err = db.Exec(`
CREATE TABLE tokens (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	text TEXT UNIQUE NOT NULL,
	is_word INTEGER NOT NULL)`)

	if err != nil {
		return err
	}

	log.Println("Creating table: token_stems")
	_, err = db.Exec(`
CREATE TABLE token_stems (
	token_id INTEGER,
	stem TEXT NOT NULL)`)

	if err != nil {
		return err
	}

	tokens := nStrings(opts.Order, func(i int) string {
		return fmt.Sprintf("token%d_id INTEGER REFERENCES token(id)", i)
	})

	log.Println("Creating table: nodes")
	_, err = db.Exec(fmt.Sprintf(`
CREATE TABLE nodes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	count INTEGER NOT NULL,
	%s)`, strings.Join(tokens, ",\n\t")))

	if err != nil {
		return err
	}

	log.Println("Creating table: edges")
	_, err = db.Exec(`
CREATE TABLE edges (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	prev_node INTEGER NOT NULL REFERENCES nodes(id),
	next_node INTEGER NOT NULL REFERENCES nodes(id),
	count INTEGER NOT NULL,
	has_space INTEGER NOT NULL)`)

	if err != nil {
		return err
	}

	db.Exec("INSERT INTO info (attribute, text) VALUES ('version', '2')")
	db.Exec("INSERT INTO info (attribute, text) VALUES ('order', ?)",
		fmt.Sprintf("%d", opts.Order))
	db.Exec("INSERT INTO info (attribute, text) VALUES ('tokenizer', ?)",
		opts.tokenizer)

	return nil
}
