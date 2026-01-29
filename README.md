# Albook - Algorithm Notebook

Albook is a lightweight, single-executable backend application with a modern web interface designed to help you track and master algorithm problems using Spaced Repetition.

## User Manual

### Getting Started
1.  **Launch**: Run `albook.exe` (or `albook` on Linux/Mac).
    - By default, it runs on port `2100`.
    - To change the port: `./albook.exe -port 8080`.
2.  **Access**: Open your browser and navigate to `http://localhost:2100`.

### Features
*   **Dashboard**:
    *   **Pending Reviews**: Count of problems due for review today or overdue.
    *   **Total Problems**: Review all submitted problems.
    *   **In Pool (Mastered)**: Problems that have passed the final review stage.
*   **Adding a Problem**:
    *   Click **"New Problem"** on the top right.
    *   Fill in Title, Source (e.g., LeetCode), ID, Resolve Date, and your Answer/Key Intuition.
    *   Optionally add a direct Link to the problem.
*   **Reviewing**:
    *   The "Problem List" shows pending items by default.
    *   Click **"Mark Reviewed"** to complete a review.
    *   **Schedule**: Reviews follow a 1-day, 3-day, 7-day, then "Pool" pattern.
    *   If a problem is not yet due, the button will show **"Wait until [Date]"**.
*   **Editing & Deleting**:
    *   Click on any problem card (not the buttons) to edit its details.
    *   To **Delete**, click the "Delete" button in the bottom-left of the edit modal.
*   **Mastered Problems**:
    *   Once a problem reaches the "Pool" stage (after 3 successful reviews), it is marked as **"Cleared"**.

## Developer Manual

### Tech Stack
*   **Language**: Go (Golang)
*   **Database**: SQLite (embedded, `albook.db` created automatically)
*   **Frontend**: Vanilla HTML/CSS/JS (embedded in binary)

### Project Structure
*   `main.go`: Entry point, HTTP server, and API handlers.
*   `db.go`: Database initialization, schema migration, and queries.
*   `static/`: Frontend assets (HTML, CSS, JS).
    *   `index.html`: Main UI layout.
    *   `style.css`: Dark mode styling and responsiveness.
    *   `app.js`: Frontend logic, API calls, state management.

### Build Instructions
Prerequisite: [Go](https://go.dev/dl/) installed.

1.  **Run Locally**:
    ```bash
    go run .
    ```
2.  **Build Executable**:
    ```bash
    go build -o albook.exe
    ```
    This creates a single independent binary file including all static assets.

### API Endpoints
*   `GET /api/dashboard`: Returns stats (pending, total, pool counts).
*   `GET /api/exercises`: List exercises (supports `?filter=pending|total|pool&page=N`).
*   `POST /api/exercises`: Create a new problem.
*   `GET /api/exercises/{id}`: Get details of a specific problem.
*   `PUT /api/exercises/{id}`: Update a problem.
*   `DELETE /api/exercises/{id}`: Delete a problem.
*   `POST /api/exercises/{id}/review`: Mark a problem as reviewed.

## Feature Roadmap

### Completed
- [x] Basic CRUD (Create, Read, Update, Delete)
- [x] Spaced Repetition Logic (1, 3, 7 days)
- [x] Dashboard Statistics
- [x] Dark Mode UI
- [x] Pagination & Filtering
- [x] Single Executable Build

### Planned (Iteration 5)
- [ ] **Pool Sampling & Review**: 
    - Allow users to click "Review" on a cleared problem.
    - System selects a problem based on "least reviewed count" and "least recently reviewed".
    - Display "Reviewed Today" count on the dashboard.
