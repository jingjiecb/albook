let state = {
    filter: 'pending',
    search: '',
    page: 1,
    totalPages: 1
};

document.addEventListener('DOMContentLoaded', () => {
    loadDashboard();
    loadExercises();

    // Event Listeners for Filters
    document.getElementById('pendingCount').parentElement.addEventListener('click', () => setFilter('pending'));
    document.getElementById('totalCount').parentElement.addEventListener('click', () => setFilter('total'));
    document.getElementById('poolCount').parentElement.addEventListener('click', () => setFilter('pool'));
    document.getElementById('reviewedTodayCount').parentElement.addEventListener('click', () => setFilter('reviewed_today'));
    document.getElementById('solvedTodayCount').parentElement.addEventListener('click', () => setFilter('solved_today'));

    // Search Listener
    const searchInput = document.getElementById('searchInput');
    if (searchInput) {
        searchInput.addEventListener('input', (e) => {
            state.search = e.target.value;
            state.page = 1;
            loadExercises();
        });
    }

    // Pagination
    document.getElementById('prevBtn').addEventListener('click', () => changePage(-1));
    document.getElementById('nextBtn').addEventListener('click', () => changePage(1));

    document.getElementById('addBtn').addEventListener('click', () => {
        openModal();
    });

    document.getElementById('cancelBtn').addEventListener('click', () => {
        document.getElementById('addModal').classList.remove('active');
    });

    document.getElementById('deleteBtn').addEventListener('click', async () => {
        const id = document.getElementById('exerciseId').value;
        if (!id || !confirm("Are you sure you want to delete this problem? This cannot be undone.")) return;

        try {
            const res = await fetch(`/api/exercises/${id}`, { method: 'DELETE' });
            if (res.ok) {
                document.getElementById('addModal').classList.remove('active');
                loadDashboard();
                loadExercises();
            }
        } catch (err) {
            console.error(err);
        }
    });

    document.getElementById('addForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        const idStr = document.getElementById('exerciseId').value;
        const id = idStr ? parseInt(idStr) : 0;

        const data = {
            title: document.getElementById('title').value,
            link: document.getElementById('link').value,
            tags: document.getElementById('tags').value,
            source: document.getElementById('source').value,
            source_id: document.getElementById('sourceId').value,
            resolve_date: new Date(document.getElementById('resolveDate').value).toISOString(),
            answer: document.getElementById('answer').value
        };

        try {
            let res;
            if (id > 0) {
                res = await fetch(`/api/exercises/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
            } else {
                res = await fetch('/api/exercises', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(data)
                });
            }

            if (res.ok) {
                document.getElementById('addModal').classList.remove('active');
                loadDashboard();
                loadExercises();
            }
        } catch (err) {
            console.error(err);
        }
    });
});

function openModal(ex = null) {
    const modal = document.getElementById('addModal');
    const form = document.getElementById('addForm');
    const deleteBtn = document.getElementById('deleteBtn');
    modal.classList.add('active');

    if (ex) {
        document.getElementById('modalTitle').textContent = "Edit Problem";
        document.getElementById('exerciseId').value = ex.id;
        document.getElementById('title').value = ex.title;
        document.getElementById('link').value = ex.link || '';
        document.getElementById('tags').value = ex.tags || '';
        document.getElementById('source').value = ex.source;
        document.getElementById('sourceId').value = ex.source_id;
        document.getElementById('resolveDate').value = new Date(ex.resolve_date).toISOString().split('T')[0];
        document.getElementById('answer').value = ex.answer;
        deleteBtn.style.display = 'block';
    } else {
        document.getElementById('modalTitle').textContent = "Add Problem";
        form.reset();
        document.getElementById('exerciseId').value = "";
        document.getElementById('resolveDate').valueAsDate = new Date();
        deleteBtn.style.display = 'none';
    }
}

async function setFilter(filter) {
    state.filter = filter;
    state.page = 1;
    updateActiveFilterUI();
    loadExercises();
}

function updateActiveFilterUI() {
    document.querySelectorAll('.stat-card').forEach(el => el.classList.remove('active'));
    if (state.filter === 'pending') document.getElementById('pendingCount').parentElement.classList.add('active');
    if (state.filter === 'total') document.getElementById('totalCount').parentElement.classList.add('active');
    if (state.filter === 'pool') document.getElementById('poolCount').parentElement.classList.add('active');
    if (state.filter === 'reviewed_today') document.getElementById('reviewedTodayCount').parentElement.classList.add('active');
    if (state.filter === 'solved_today') document.getElementById('solvedTodayCount').parentElement.classList.add('active');
}

async function changePage(delta) {
    const newPage = state.page + delta;
    if (newPage > 0 && newPage <= state.totalPages) {
        state.page = newPage;
        loadExercises();
    }
}

async function loadDashboard() {
    try {
        const res = await fetch('/api/dashboard');
        const data = await res.json();

        document.getElementById('pendingCount').textContent = data.pending_count;
        document.getElementById('totalCount').textContent = data.total_count;
        document.getElementById('poolCount').textContent = data.pool_count;
        document.getElementById('reviewedTodayCount').textContent = data.reviewed_today_count;
        document.getElementById('solvedTodayCount').textContent = data.solved_today_count;

        // Initial active state
        updateActiveFilterUI();
    } catch (err) {
        console.error("Failed to load dashboard", err);
    }
}

async function loadExercises() {
    try {
        const res = await fetch(`/api/exercises?filter=${state.filter}&page=${state.page}&search=${encodeURIComponent(state.search)}`);
        const data = await res.json();

        state.totalPages = data.total_pages;
        document.getElementById('pageInfo').textContent = `Page ${state.page} of ${state.totalPages}`;
        document.getElementById('prevBtn').disabled = state.page === 1;
        document.getElementById('nextBtn').disabled = state.page === state.totalPages;

        const list = document.getElementById('pendingList');
        list.innerHTML = '';

        if (!data.data || data.data.length === 0) {
            list.innerHTML = '<div class="empty-state">No problems found for this filter.</div>';
            return;
        }

        data.data.forEach(ex => {
            const div = document.createElement('div');
            div.className = 'exercise-card';
            div.style.cursor = 'pointer';

            // Click to Edit
            div.addEventListener('click', (e) => {
                // Prevent open if clicking link or review button
                if (e.target.tagName === 'A' || e.target.classList.contains('review-btn')) return;
                openModal(ex);
            });

            let linkHtml = '';
            if (ex.link) {
                linkHtml = `<a href="${ex.link}" target="_blank" class="link-btn">Open Link</a>`;
            }

            // Specific UI tweaks for Pool view?
            // Maybe hide "Mark Reviewed" if in pool? User didn't specify, but implies "Mastered".
            // However, user said "then put it into a exercise pool". We can still review them if we want?
            // I'll keep the review button available everywhere for now, 
            // but maybe change text if it's already in pool.

            // Review Availability Logic
            const nextReview = new Date(ex.next_review_date);
            const now = new Date();
            const isPending = nextReview <= now; // Should correspond to backend logic roughly, or just check date

            let buttonHtml = '';
            // If strictly enforcing "cannot review until", we check dates.
            // Note: DB "pending" filter already filters by date.
            // But if user clicks "Total", they see future items.

            if (isPending && ex.review_stage < 3) {
                buttonHtml = `<button class="review-btn" onclick="markReviewed(event, ${ex.id})">Mark Reviewed</button>`;
            } else if (ex.review_stage >= 3) {
                buttonHtml = `<button class="review-btn" disabled>Cleared</button>`;
            } else {
                // Future date
                buttonHtml = `<button class="review-btn" disabled>Wait until ${nextReview.toLocaleDateString()}</button>`;
            }

            div.innerHTML = `
                <div class="exercise-info">
                    <h3>${esc(ex.title)}</h3>
                    <div class="exercise-meta">
                        <span class="badge">${esc(ex.source)} ${esc(ex.source_id)}</span>
                        ${ex.tags ? `<span class="badge" style="background:var(--secondary-bg); color:var(--text);">${esc(ex.tags)}</span>` : ''}
                        <span>Reviews: ${ex.review_count}</span>
                    </div>
                </div>
                <div class="exercise-actions">
                    ${linkHtml}
                    ${buttonHtml}
                </div>
            `;
            list.appendChild(div);
        });
    } catch (err) {
        console.error("Failed to load exercises", err);
    }
}



async function markReviewed(event, id) {
    if (event) event.stopPropagation();
    if (!confirm("Confirm you have reviewed this problem?")) return;
    try {
        const res = await fetch(`/api/exercises/${id}/review`, { method: 'POST' });
        if (res.ok) {
            loadDashboard(); // Refresh stats
            loadExercises(); // Refresh list
        }
    } catch (err) {
        alert("Error reviewing: " + err);
    }
}

function esc(str) {
    if (!str) return '';
    const div = document.createElement('div');
    div.innerText = str;
    return div.innerHTML;
}
