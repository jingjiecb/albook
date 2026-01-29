package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS exercises (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source TEXT,
		source_id TEXT,
		title TEXT NOT NULL,
		resolve_date DATETIME NOT NULL,
		next_review_date DATETIME NOT NULL,
		review_stage INTEGER DEFAULT 0,
		review_count INTEGER DEFAULT 0,
		answer TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = DB.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}

	// Migration: Add review_count if not exists
	_, _ = DB.Exec("ALTER TABLE exercises ADD COLUMN review_count INTEGER DEFAULT 0")
	// Migration: Add link if not exists
	_, _ = DB.Exec("ALTER TABLE exercises ADD COLUMN link TEXT")
}

type Exercise struct {
	ID             int       `json:"id"`
	Source         string    `json:"source"`
	SourceID       string    `json:"source_id"`
	Title          string    `json:"title"`
	Link           string    `json:"link"`
	ResolveDate    time.Time `json:"resolve_date"`
	NextReviewDate time.Time `json:"next_review_date"`
	ReviewStage    int       `json:"review_stage"`
	ReviewCount    int       `json:"review_count"`
	Answer         string    `json:"answer"`
	CreatedAt      time.Time `json:"created_at"`
}

func CreateExercise(e Exercise) (int64, error) {
	// Calculate initial next_review_date based on resolve_date
	// Stage 0 -> 1 day after resolve date
	e.NextReviewDate = e.ResolveDate.AddDate(0, 0, 1)
	e.ReviewStage = 0
	e.ReviewCount = 0

	stmt, err := DB.Prepare("INSERT INTO exercises(source, source_id, title, link, resolve_date, next_review_date, review_stage, review_count, answer) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(e.Source, e.SourceID, e.Title, e.Link, e.ResolveDate, e.NextReviewDate, e.ReviewStage, e.ReviewCount, e.Answer)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetExercises(filter string, page int, pageSize int) ([]Exercise, int, error) {
	offset := (page - 1) * pageSize

	baseQuery := "SELECT id, source, source_id, title, IFNULL(link, ''), resolve_date, next_review_date, review_stage, review_count, answer, created_at FROM exercises"
	countQuery := "SELECT COUNT(*) FROM exercises"
	whereClause := ""

	switch filter {
	case "pending":
		whereClause = " WHERE next_review_date <= datetime('now') AND review_stage < 3"
	case "pool":
		whereClause = " WHERE review_stage >= 3"
	case "total":
		whereClause = ""
	default: // Default to pending if unknown, or maybe all? logic says pending is default view.
		whereClause = " WHERE next_review_date <= datetime('now') AND review_stage < 3"
	}

	orderBy := " ORDER BY next_review_date ASC"
	if filter == "total" || filter == "pool" {
		orderBy = " ORDER BY created_at DESC"
	}

	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	// Get Count
	var totalItems int
	err := DB.QueryRow(countQuery + whereClause).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	// Get Data
	rows, err := DB.Query(baseQuery + whereClause + orderBy + limitClause)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var e Exercise
		err = rows.Scan(&e.ID, &e.Source, &e.SourceID, &e.Title, &e.Link, &e.ResolveDate, &e.NextReviewDate, &e.ReviewStage, &e.ReviewCount, &e.Answer, &e.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		exercises = append(exercises, e)
	}
	return exercises, totalItems, nil
}

func GetExerciseByID(id int) (*Exercise, error) {
	var e Exercise
	err := DB.QueryRow("SELECT id, source, source_id, title, IFNULL(link, ''), resolve_date, next_review_date, review_stage, review_count, answer, created_at FROM exercises WHERE id = ?", id).
		Scan(&e.ID, &e.Source, &e.SourceID, &e.Title, &e.Link, &e.ResolveDate, &e.NextReviewDate, &e.ReviewStage, &e.ReviewCount, &e.Answer, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func UpdateExercise(e Exercise) error {
	// Only update info fields, not review progress
	_, err := DB.Exec("UPDATE exercises SET source=?, source_id=?, title=?, link=?, resolve_date=?, answer=? WHERE id=?",
		e.Source, e.SourceID, e.Title, e.Link, e.ResolveDate, e.Answer, e.ID)
	return err
}

func DeleteExercise(id int) error {
	_, err := DB.Exec("DELETE FROM exercises WHERE id = ?", id)
	return err
}

func GetStats() (int, int, error) {
	var total int
	var pool int

	err := DB.QueryRow("SELECT COUNT(*) FROM exercises").Scan(&total)
	if err != nil {
		return 0, 0, err
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM exercises WHERE review_stage >= 3").Scan(&pool)
	if err != nil {
		return 0, 0, err
	}

	return total, pool, nil
}

func PerformReview(id int) error {
	// Fetch current stage and review count
	var stage int
	var count int
	err := DB.QueryRow("SELECT review_stage, IFNULL(review_count, 0) FROM exercises WHERE id = ?", id).Scan(&stage, &count)
	if err != nil {
		return err
	}

	// Interval logic: 1, 3, 7 days.
	// Stage 0 (First submit) -> +1 day -> Review 1 (Stage 1)
	// Review 1 (Stage 1) -> +3 days -> Review 2 (Stage 2)
	// Review 2 (Stage 2) -> +7 days -> Review 3 (Stage 3/Pool)

	// If currently at stage 0 (meaning waiting for 1st review), and we review it:
	// New Stage = 1. Next Date = Now + 3 days?
	// User said: "1 day means after 1 day from the initial resolve date, the system should remind the user it needs reviewing, and 3 days means after the first review, the system should reminder the user to review it 3 days later from the first review day"

	newStage := stage + 1
	newCount := count + 1
	daysToAdd := 0

	switch newStage {
	case 1:
		daysToAdd = 3
	case 2:
		daysToAdd = 7
	default:
		// Pool - 3
		// Check if it's already in pool (>=3)
		if stage >= 3 {
			// If reviewing something already in pool, just keep it there, maybe add 30 days?
			// For now, let's just increment count and keep in pool.
			newStage = 3   // Cap at 3 for "Pool" state definition
			daysToAdd = 30 // Periodic review for pool? Or just far future? User said "put it into exercise pool"
		} else {
			daysToAdd = 36500 // Basically done / in pool
		}
	}

	now := time.Now()
	nextDue := now.AddDate(0, 0, daysToAdd)

	_, err = DB.Exec("UPDATE exercises SET review_stage = ?, review_count = ?, next_review_date = ? WHERE id = ?", newStage, newCount, nextDue, id)
	return err
}
