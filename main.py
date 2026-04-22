from fastapi import FastAPI
from memory_graph import MemoryGraph

app = FastAPI()
memory = MemoryGraph()

@app.get("/")
async def health_check():
    return {"status": "ok"}

@app.post("/analyze-meeting")
async def analyze_meeting(transcript: str):
    # Process transcript (stub for now)
    meeting_id = memory.add_meeting(transcript)
    return {"meeting_id": meeting_id}

@app.get("/memory/meetings")
async def get_meetings():
    return {"meetings": memory.get_all_meetings()}
