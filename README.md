# Albook - Algorithm Notebook

Albook is a lightweight, single-executable backend application with a modern web interface designed to help you track and master algorithm problems using Spaced Repetition.

## User Manual

### Getting Started
1.  **Launch**: Run `albook.exe` (or `albook` on Linux/Mac).
    - By default, it runs on port `2100`.
    - To change the port: `./albook.exe -port 8080`.
    - To specify a database file: `./albook.exe -db ./my_data.db` (Default: `./albook.db`).
    - If the database file does not exist, a new one will be initialized with a warning message.
2.  **Access**: Open your browser and navigate to `http://localhost:2100`.

### Features
*   **Dashboard**:
    *   **Pending Reviews**: Count of problems due for review today or overdue.
    *   **Solved Today**: Count of problems resolved in the current day.
    *   **Reviewed Today**: Count of problems reviewed in the current day.
    *   **Total Problems**: Review all submitted problems.
    *   **In Pool (Mastered)**: Problems that have passed the final review stage.
*   **Adding a Problem**:
    *   Click **"New Problem"** on the top right.
    *   Fill in Title, Source (e.g., LeetCode), ID, Resolve Date, and your Answer/Key Intuition.
    *   **Tags**: Add relevant tags (e.g., "DP", "Greedy") to categorize the problem.
    *   Optionally add a direct Link to the problem.
*   **Search**:
    *   Use the search box on the main page to filter problems by Title, ID, Tags, or Description.
*   **Reviewing**:
    *   The "Problem List" shows pending items by default.
    *   Click **"Mark Reviewed"** to complete a review.
    *   **Schedule**: Reviews follow a 1-day, 3-day, 7-day, then "Pool" pattern.
    *   If a problem is not yet due, the button will show **"Wait until [Date]"**.
*   **Editing & Deleting**:
    *   Click on any problem card (not the buttons) to edit its details.
    *   To **Delete**, click the "Delete" button in the bottom-left of the edit modal. This moves the problem to Trash.
*   **Trash**:
    *   The **Trash** tab on the dashboard shows how many problems are in the trash.
    *   Click the Trash tab to view trashed problems.
    *   From the trash view, you can **Restore** a problem back to your active set or **Permanently Delete** it.
    *   You can also click on a trashed problem to view its details.
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
    ```
    This creates a single independent binary file including all static assets.

### Docker Deployment

**Prerequisites**: Docker and Docker Compose installed.

1.  **Set Secrets (GitHub Actions)**:
    For automatic building and pushing, go to your GitHub Repo -> Settings -> Secrets and Variables -> Actions, and add:
    *   `DOCKER_USERNAME`: Your Docker Hub username.
    *   `DOCKER_PASSWORD`: Your Docker Hub access token (preferred) or password.

    Once set, pushing to `master`/`main` triggers an `edge` tag build. Pushing a tag like `v1.0.0` triggers a versioned build.

2.  **Run with Docker Compose**:
    Creates a container running on port 2100 with persistent data in `./data`.

    ```bash
    # Set the image name (replace 'yourusername' with your actual username)
    export DOCKER_IMAGE_NAME=yourusername/albook
    
    # Run
    docker-compose up -d
    ```

    Alternatively, edit `docker-compose.yml` directly.
*   `GET /api/dashboard`: Returns stats (pending, total, pool, trash counts).
*   `GET /api/problems`: List problems (supports `?filter=pending|total|trash&page=N&search=KEYWORD`).
*   `POST /api/problems`: Create a new problem.
*   `GET /api/problems/{id}`: Get details of a specific problem.
*   `PUT /api/problems/{id}`: Update a problem.
*   `DELETE /api/problems/{id}`: Soft-delete (move to trash) or permanently delete (if already in trash).
*   `POST /api/problems/{id}/review`: Mark a problem as reviewed.
*   `PUT /api/problems/{id}/restore`: Restore a trashed problem.

## Feature Roadmap

### Completed
- [x] Basic CRUD (Create, Read, Update, Delete)
- [x] Spaced Repetition Logic (1, 3, 7 days)
- [x] Dashboard Statistics
- [x] Dark Mode UI
- [x] Pagination & Filtering
- [x] Search (Title, Tags, ID, Description)
- [x] Tags Support
- [x] Single Executable Build
- [x] Add "Reviewed Today" and "Solved Today" tabs.
- [x] **Pool Review**: 
    - Change cleared to a little green badge with "cleared" in it, instead of change the review button to cleared.
    - Allow users to click "Review" on a cleared problem, then the problem's review count will be further incremented.
    - Delete the current pool tab, and rename current "total problems" to "problem pool".
    - Make sure the new problem pool tab sorts the problem by 1) review count from less to more 2) last review date(time) from new to old.
- [x] **Display last review data**:
    - Add the last review date on the problem card.
- [x] **Trash Bin**:
    - Soft-delete: deleting a problem moves it to trash instead of permanently removing it.
    - Restore: trashed problems can be restored back to the active set.
    - Permanent delete: delete from trash to permanently remove.