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
	// Migration: Add tags if not exists
	_, _ = DB.Exec("ALTER TABLE exercises ADD COLUMN tags TEXT")
	// Migration: Add last_reviewed_at if not exists
	_, _ = DB.Exec("ALTER TABLE exercises ADD COLUMN last_reviewed_at DATETIME")
}

type Exercise struct {
	ID             int       `json:"id"`
	Source         string    `json:"source"`
	SourceID       string    `json:"source_id"`
	Title          string    `json:"title"`
	Link           string    `json:"link"`
	Tags           string    `json:"tags"`
	ResolveDate    time.Time `json:"resolve_date"`
	NextReviewDate time.Time `json:"next_review_date"`
	ReviewStage    int       `json:"review_stage"`
	ReviewCount    int       `json:"review_count"`
	Answer         string    `json:"answer"`
	CreatedAt      time.Time `json:"created_at"`
	LastReviewedAt time.Time `json:"last_reviewed_at"`
}

func CreateExercise(e Exercise) (int64, error) {
	// Calculate initial next_review_date based on resolve_date
	// Stage 0 -> 1 day after resolve date
	e.NextReviewDate = e.ResolveDate.AddDate(0, 0, 1)
	e.ReviewStage = 0
	e.ReviewCount = 0

	stmt, err := DB.Prepare("INSERT INTO exercises(source, source_id, title, link, tags, resolve_date, next_review_date, review_stage, review_count, answer) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	res, err := stmt.Exec(e.Source, e.SourceID, e.Title, e.Link, e.Tags, e.ResolveDate, e.NextReviewDate, e.ReviewStage, e.ReviewCount, e.Answer)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func GetExercises(filter string, search string, page int, pageSize int) ([]Exercise, int, error) {
	offset := (page - 1) * pageSize

	baseQuery := "SELECT id, source, source_id, title, IFNULL(link, ''), IFNULL(tags, ''), resolve_date, next_review_date, review_stage, review_count, answer, created_at FROM exercises"
	countQuery := "SELECT COUNT(*) FROM exercises"

	whereClause := " WHERE 1=1" // Base where for easier appending

	// Apply Filter
	switch filter {
	case "pending":
		whereClause += " AND next_review_date <= datetime('now') AND review_stage < 3"
	case "pool":
		whereClause += " AND review_stage >= 3"
	case "reviewed_today":
		whereClause += " AND date(last_reviewed_at) = date('now', 'localtime')"
	case "total":
		// No extra filter
	default:
		whereClause += " AND next_review_date <= datetime('now') AND review_stage < 3"
	}

	// Apply Search
	var args []interface{}
	if search != "" {
		searchLike := "%" + search + "%"
		// Search in source_id, title, tags, answer
		whereClause += " AND (source_id LIKE ? OR title LIKE ? OR tags LIKE ? OR answer LIKE ?)"
		// We need to pass args to Query.
		// Note: Count query also needs these args.
		args = append(args, searchLike, searchLike, searchLike, searchLike)
	}

	orderBy := " ORDER BY next_review_date ASC"
	if filter == "total" || filter == "pool" || filter == "reviewed_today" {
		orderBy = " ORDER BY created_at DESC"
	}

	limitClause := fmt.Sprintf(" LIMIT %d OFFSET %d", pageSize, offset)

	// Get Count
	var totalItems int
	err := DB.QueryRow(countQuery+whereClause, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	// Get Data
	rows, err := DB.Query(baseQuery+whereClause+orderBy+limitClause, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var exercises []Exercise
	for rows.Next() {
		var e Exercise
		err = rows.Scan(&e.ID, &e.Source, &e.SourceID, &e.Title, &e.Link, &e.Tags, &e.ResolveDate, &e.NextReviewDate, &e.ReviewStage, &e.ReviewCount, &e.Answer, &e.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		exercises = append(exercises, e)
	}
	return exercises, totalItems, nil
}

func GetExerciseByID(id int) (*Exercise, error) {
	var e Exercise
	err := DB.QueryRow("SELECT id, source, source_id, title, IFNULL(link, ''), IFNULL(tags, ''), resolve_date, next_review_date, review_stage, review_count, answer, created_at FROM exercises WHERE id = ?", id).
		Scan(&e.ID, &e.Source, &e.SourceID, &e.Title, &e.Link, &e.Tags, &e.ResolveDate, &e.NextReviewDate, &e.ReviewStage, &e.ReviewCount, &e.Answer, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func UpdateExercise(e Exercise) error {
	// Only update info fields, not review progress
	_, err := DB.Exec("UPDATE exercises SET source=?, source_id=?, title=?, link=?, tags=?, resolve_date=?, answer=? WHERE id=?",
		e.Source, e.SourceID, e.Title, e.Link, e.Tags, e.ResolveDate, e.Answer, e.ID)
	return err
}

func DeleteExercise(id int) error {
	_, err := DB.Exec("DELETE FROM exercises WHERE id = ?", id)
	return err
}

// ReviewIntervals defines days to add for each stage
var ReviewIntervals = map[int]int{
	1: 3,
	2: 7,
	3: 30, // Pool interval
}

func GetStats() (int, int, int, int, error) {
	var total int
	var pool int
	var pending int
	var reviewedToday int

	query := `
		SELECT 
			COUNT(*),
			SUM(CASE WHEN review_stage >= 3 THEN 1 ELSE 0 END),
			SUM(CASE WHEN next_review_date <= datetime('now') AND review_stage < 3 THEN 1 ELSE 0 END),
			SUM(CASE WHEN date(last_reviewed_at) = date('now', 'localtime') THEN 1 ELSE 0 END)
		FROM exercises
	`

	err := DB.QueryRow(query).Scan(&total, &pool, &pending, &reviewedToday)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return total, pool, pending, reviewedToday, nil
}

func PerformReview(id int) error {
	// Fetch current stage and review count
	var stage int
	var count int
	err := DB.QueryRow("SELECT review_stage, IFNULL(review_count, 0) FROM exercises WHERE id = ?", id).Scan(&stage, &count)
	if err != nil {
		return err
	}

	newStage := stage + 1
	newCount := count + 1
	daysToAdd := 0

	if val, ok := ReviewIntervals[newStage]; ok {
		daysToAdd = val
	} else {
		if newStage > 3 {
			// Already in pool, keep adding 30 days or handle as mastered
			newStage = 3
			daysToAdd = 30
		} else {
			// Fallback (e.g. stage 0 which shouldn't happen here usually if logic aligns)
			// But logic says stage 0 -> 1.
			// If newStage is 1, it's in map.
			daysToAdd = 365
		}
	}

	now := time.Now()
	nextDue := now.AddDate(0, 0, daysToAdd)

	_, err = DB.Exec("UPDATE exercises SET review_stage = ?, review_count = ?, next_review_date = ?, last_reviewed_at = ? WHERE id = ?", newStage, newCount, nextDue, now, id)
	return err
}
