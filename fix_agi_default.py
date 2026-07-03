with open('agents/swarm_manager.go', 'r') as f:
    content = f.read()

old_new = """func NewSwarmManager() *SwarmManager {
	sm := &SwarmManager{
		Swarms:         make(map[string][]SpecialistAgent),
		Factories:      make(map[string]func() []SpecialistAgent),
		LongTermMemory: make(map[string]string),
		MemoryPath:     "./data/agi_memory.json",
		TaskForces:     make(map[string]*TaskForce),
	}"""

new_new = """func NewSwarmManager() *SwarmManager {
	sm := &SwarmManager{
		Swarms:         make(map[string][]SpecialistAgent),
		Factories:      make(map[string]func() []SpecialistAgent),
		LongTermMemory: make(map[string]string),
		MemoryPath:     "./data/agi_memory.json",
		TaskForces:     make(map[string]*TaskForce),
		AGIMode:        true,
	}"""

content = content.replace(old_new, new_new)

with open('agents/swarm_manager.go', 'w') as f:
    f.write(content)
