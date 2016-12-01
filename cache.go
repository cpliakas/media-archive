package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type CacheItem struct {
	Filename string
}

func (i CacheItem) String() string {
	return i.Filename
}

type Cache interface {

	// Set adds an item to the cache.
	Set(CacheItem) error

	// Get returns a cache item by unique key.
	Get(string) (CacheItem, error)

	// Purge removes an item from the cache.
	Purge(string) error
}

type SQLiteCache struct {
	db      *sql.DB
	archive string
}

func NewSQLiteCache(archive string) (Cache, error) {
	basedir := fmt.Sprintf("/Users/%s/Library/Caches/com.chrispliakas.media-archive", os.Getenv("USER"))
	cachefile := fmt.Sprintf("%s/cache.db", basedir)

	// Ensure the cache directory is available.
	stat, err := os.Stat(basedir)
	if os.IsNotExist(err) {
		if err := os.Mkdir(basedir, 0755); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, err
	}

	// Open the database.
	db, err := sql.Open("sqlite3", cachefile)
	if err != nil {
		return nil, err
	}

	sql := `
		CREATE TABLE IF NOT EXISTS %s (
    		filename VARCHAR(255) PRIMARY KEY
		);
	`
	_, err = db.Exec(fmt.Sprintf(sql, "`"+archive+"`"))
	if err != nil {
		return nil, err
	}

	return &SQLiteCache{db: db, archive: archive}, nil
}

func (c *SQLiteCache) Set(item CacheItem) (err error) {
	stmt, err := c.db.Prepare(fmt.Sprintf("INSERT INTO `%s`(filename) values(?)", c.archive))
	if err != nil {
		return
	}

	_, err = stmt.Exec(item.Filename)
	return
}

func (c *SQLiteCache) Get(filename string) (item CacheItem, err error) {
	sql := fmt.Sprintf("SELECT filename FROM `%s` WHERE filename = ?", c.archive)
	rows, err := c.db.Query(sql, filename)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		item = CacheItem{}
		err = rows.Scan(&item.Filename)
		return
	}

	return
}

func (c *SQLiteCache) Purge(filename string) (err error) {
	stmt, err := c.db.Prepare(fmt.Sprintf("DELETE FROM `%s` WHERE filename = ?", c.archive))
	if err != nil {
		return
	}

	_, err = stmt.Exec(filename)
	return
}
