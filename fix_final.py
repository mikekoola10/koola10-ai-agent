with open('agents/swarm_manager.go', 'r') as f:
    lines = f.readlines()

new_lines = []
in_new = False
for line in lines:
    if 'func NewSwarmManager() *SwarmManager {' in line:
        in_new = True
        new_lines.append(line)
        new_lines.append('\tsm := &SwarmManager{\n')
        new_lines.append('\t\tSwarms:         make(map[string][]SpecialistAgent),\n')
        new_lines.append('\t\tFactories:      make(map[string]func() []SpecialistAgent),\n')
        new_lines.append('\t\tLongTermMemory: make(map[string]string),\n')
        new_lines.append('\t\tMemoryPath:     "./data/agi_memory.json",\n')
        new_lines.append('\t\tTaskForces:     make(map[string]*TaskForce),\n')
        new_lines.append('\t\tAGIMode:        true,\n')
        new_lines.append('\t}\n')
        continue
    if in_new:
        if 'sm.LoadMemory()' in line:
            in_new = False
            new_lines.append(line)
        continue
    new_lines.append(line)

with open('agents/swarm_manager.go', 'w') as f:
    f.writelines(new_lines)

with open('dashboard.html', 'r') as f:
    db = f.read()

if 'fetch(\'/admin/agi-mode\')' not in db:
    db = db.replace('const eventSource = new EventSource(\'/events/stream\');', """
        // Sync AGI Mode state on load
        fetch('/admin/agi-mode', {
            headers: { 'X-Admin-API-Key': 'MzE5OGYzNGEtZmM1ZC00YjY3LWI3ZGMtYjZiOTc5YzdjNzUyYjcwNDczMjYtNjg4Yi00OGIzLTg3NzMtZGQzOTc5NTViZmE0' }
        })
        .then(res => res.json())
        .then(data => {
            agiEnabled = data.enabled;
            updateAGIUI();
        });

        const eventSource = new EventSource('/events/stream');""")

with open('dashboard.html', 'w') as f:
    f.write(db)
