package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func parseTime(val interface{}) (int64, error) {
	if val == nil {
		return 0, fmt.Errorf("nil value")
	}

	switch v := val.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	case []byte:
		return parseStringTime(string(v))
	case string:
		return parseStringTime(v)
	case time.Time:
		return v.Unix(), nil
	default:
		return 0, fmt.Errorf("unknown type: %T", v)
	}
}

func parseStringTime(s string) (int64, error) {
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.Unix(), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.Unix(), nil
	}
	// Try SQLite default format "2006-01-02 15:04:05" (assumed UTC for fallback)
	if t, err := time.Parse("2006-01-02 15:04:05", s); err == nil {
		return t.Unix(), nil
	}
	// Try generic ISO8601 without timezone
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t.Unix(), nil
	}
	// Another format possibly found in driver
	if t, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", s); err == nil {
		return t.Unix(), nil
	}
	
	return 0, fmt.Errorf("unable to parse time string")
}

func main() {
	dbPath := flag.String("db", "./albook.db", "Path to the SQLite database file")
	flag.Parse()

	fmt.Printf("Starting migration to 10-digit Unix timestamps for database: %s\n", *dbPath)
	
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, resolve_date, next_review_date, created_at, last_reviewed_at FROM exercises")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	updateStmt, err := tx.Prepare("UPDATE exercises SET resolve_date=?, next_review_date=?, created_at=?, last_reviewed_at=? WHERE id=?")
	if err != nil {
		log.Fatal(err)
	}
	defer updateStmt.Close()

	updatedCount := 0

	for rows.Next() {
		var id int
		var resolve, next, created, lastRev interface{}
		err := rows.Scan(&id, &resolve, &next, &created, &lastRev)
		if err != nil {
			log.Fatalf("Error scanning row %d: %v", id, err)
		}

		resolveInt, _ := parseTime(resolve)
		nextInt, _ := parseTime(next)

		var createdInt interface{}
		if c, err := parseTime(created); err == nil {
			createdInt = c
		} else {
			createdInt = nil // keeping it null if missing or error
		}

		var lastRevInt interface{}
		if l, err := parseTime(lastRev); err == nil {
			lastRevInt = l
		} else {
			lastRevInt = nil
		}

		_, err = updateStmt.Exec(resolveInt, nextInt, createdInt, lastRevInt, id)
		if err != nil {
			log.Fatalf("Error updating row %d: %v", id, err)
		}
		updatedCount++
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Successfully migrated %d records to 10-digit unix timestamps.\n", updatedCount)
}
