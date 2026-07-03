with open('dashboard.html', 'r') as f:
    db = f.read()

# Fix the SSE connection status UI update
old_sse = """        const eventSource = new EventSource('/events/stream');
        eventSource.onmessage = (e) => {
            const data = JSON.parse(e.data);
            if (data.type === 'log') {
                updateLogs(data.content);
            }
        };"""

new_sse = """        const eventSource = new EventSource('/events/stream');
        eventSource.onopen = () => {
            document.getElementById('status-dot').style.backgroundColor = '#00ff00';
            document.getElementById('status-text').innerText = 'CONNECTED';
            document.getElementById('status-text').style.color = '#00ff00';
        };
        eventSource.onmessage = (e) => {
            const data = JSON.parse(e.data);
            if (data.type === 'log') {
                updateLogs(data.content);
            }
            if (data.type === 'coordination') {
                updateCoordination(data.content);
            }
        };
        eventSource.onerror = () => {
            document.getElementById('status-dot').style.backgroundColor = '#ff0000';
            document.getElementById('status-text').innerText = 'CONNECTION_LOST';
            document.getElementById('status-text').style.color = '#ff0000';
        };"""

db = db.replace(old_sse, new_sse)

with open('dashboard.html', 'w') as f:
    f.write(db)
