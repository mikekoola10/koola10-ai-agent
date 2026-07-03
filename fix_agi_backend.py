import sys

with open('main.go', 'r') as f:
    content = f.read()

# 1. Update handleEventsStream to keep connection alive
old_stream = """func handleEventsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Write([]byte("event: connected\\ndata: {}\\n\\n"))
}"""

new_stream = """func handleEventsStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("event: connected\\ndata: {}\\n\\n"))
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.Write([]byte(": keepalive\\n\\n"))
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}"""

content = content.replace(old_stream, new_stream)

# 2. Add /admin/agi-mode GET endpoint
if 'r.Get("/admin/agi-mode"' not in content:
    content = content.replace('r.Post("/admin/agi-mode",', 'r.Get("/admin/agi-mode", corsMiddleware(authMiddleware(handleGetAGIMode)))\n\tr.Post("/admin/agi-mode",')

# 3. Add handleGetAGIMode function
if 'func handleGetAGIMode' not in content:
    content += """
func handleGetAGIMode(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"enabled": globalSwarmManager.IsAGIMode()})
}
"""

with open('main.go', 'w') as f:
    f.write(content)

with open('dashboard.html', 'r') as f:
    dashboard = f.read()

# Update toggle to sync with server on load
sync_code = """
        // Sync AGI Mode state on load
        fetch('/admin/agi-mode', {
            headers: { 'X-Admin-API-Key': 'MzE5OGYzNGEtZmM1ZC00YjY3LWI3ZGMtYjZiOTc5YzdjNzUyYjcwNDczMjYtNjg4Yi00OGIzLTg3NzMtZGQzOTc5NTViZmE0' }
        })
        .then(res => res.json())
        .then(data => {
            agiEnabled = data.enabled;
            updateAGIUI();
        });
"""

if 'fetch(\'/admin/agi-mode\'' not in dashboard:
    dashboard = dashboard.replace('const eventSource = new EventSource(\'/events/stream\');', sync_code + '\n        const eventSource = new EventSource(\'/events/stream\');')

with open('dashboard.html', 'w') as f:
    f.write(dashboard)
