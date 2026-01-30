package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	dbPath := flag.String("db", "./albook.db", "Path to the database file")
	port := flag.Int("port", 2100, "Port for the web server")
	flag.Parse()

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		fmt.Printf("Warning: Database file '%s' does not exist. Creating a new one.\n", *dbPath)
	}

	InitDB(*dbPath)

	http.HandleFunc("GET /api/dashboard", handleDashboard)
	http.HandleFunc("GET /api/exercises", handleListExercises)
	http.HandleFunc("POST /api/exercises", handleCreateExercise)
	http.HandleFunc("GET /api/exercises/{id}", handleGetExercise)
	http.HandleFunc("PUT /api/exercises/{id}", handleUpdateExercise)
	http.HandleFunc("DELETE /api/exercises/{id}", handleDeleteExercise)
	http.HandleFunc("POST /api/exercises/{id}/review", handleReviewExercise)

	// Static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/", http.FileServer(http.FS(staticFS)))

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Using database: %s\n", *dbPath)
	fmt.Printf("Server starting on http://localhost%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	total, pool, pending, err := GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type DashboardResponse struct {
		PendingCount int `json:"pending_count"`
		TotalCount   int `json:"total_count"`
		PoolCount    int `json:"pool_count"`
	}

	resp := DashboardResponse{
		PendingCount: pending,
		TotalCount:   total,
		PoolCount:    pool,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleListExercises(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "pending"
	}

	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err == nil && p > 0 {
			page = p
		}
	}

	search := r.URL.Query().Get("search")

	pageSize := 10

	exercises, total, err := GetExercises(filter, search, page, pageSize)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	type ListResponse struct {
		Data       []Exercise `json:"data"`
		Total      int        `json:"total"`
		Page       int        `json:"page"`
		TotalPages int        `json:"total_pages"`
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	resp := ListResponse{
		Data:       exercises,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleCreateExercise(w http.ResponseWriter, r *http.Request) {
	var e Exercise
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if e.ResolveDate.IsZero() {
		e.ResolveDate = time.Now()
	}

	id, err := CreateExercise(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func handleGetExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}
	e, err := GetExerciseByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(e)
}

func handleUpdateExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var e Exercise
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	e.ID = id

	if err := UpdateExercise(e); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func handleDeleteExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := DeleteExercise(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "deleted"}`))
}

func handleReviewExercise(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err := PerformReview(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "reviewed"}`))
}
