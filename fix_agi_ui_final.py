with open('dashboard.html', 'r') as f:
    db = f.read()

# Fix initialization logic
init_logic = """
        let agiEnabled = true;

        async function syncState() {
            try {
                const res = await fetch('/admin/agi-mode', {
                    headers: { 'X-Admin-API-Key': 'MzE5OGYzNGEtZmM1ZC00YjY3LWI3ZGMtYjZiOTc5YzdjNzUyYjcwNDczMjYtNjg4Yi00OGIzLTg3NzMtZGQzOTc5NTViZmE0' }
                });
                const data = await res.json();
                agiEnabled = data.enabled;
                updateAGIUI();
            } catch (e) {
                console.error('Failed to sync state:', e);
            }
        }

        function updateAGIUI() {
            document.getElementById('agi-status').innerText = agiEnabled ? 'ACTIVE' : 'OFF';
            document.getElementById('agi-toggle-btn').innerText = agiEnabled ? 'DISABLE_AGI_MODE' : 'ENABLE_AGI_MODE';
            document.getElementById('agi-status').style.color = agiEnabled ? '#00ff00' : '#ff0000';
        }

        // Initialize
        syncState();
        updateAGIUI();

        const eventSource = new EventSource('/events/stream');
"""

# Replace old init and EventSource setup
import re
db = re.sub(r'let agiEnabled = (true|false);.*?const eventSource = new EventSource\(\'/events/stream\'\);', init_logic, db, flags=re.DOTALL)

with open('dashboard.html', 'w') as f:
    f.write(db)
