from fastapi import FastAPI, Body, Query, HTTPException
from memory_graph import MemoryGraph
from tools import registry
from semantic_memory import SemanticMemory
from agents import Orchestrator
from safety import ControlPlane

app = FastAPI()
memory = MemoryGraph()
semantic = SemanticMemory()
orchestrator = Orchestrator()
control_plane = ControlPlane(threshold=0.6)

@app.get("/")
async def health_check():
    return {"status": "ok"}

@app.post("/analyze-meeting")
async def analyze_meeting(transcript: str):
    # Process transcript (stub for now)
    meeting_id = memory.add_meeting(transcript)

    # Index into semantic memory
    semantic.add_text(transcript, meeting_id, {"type": "transcript"})
    # Stub for action items
    semantic.add_text("Action Item: Stub item for " + meeting_id, meeting_id, {"type": "action_item"})

    return {"meeting_id": meeting_id}

@app.get("/memory/meetings")
async def get_meetings():
    return {"meetings": memory.get_all_meetings()}

@app.post("/tools/execute")
async def execute_tool(tool_name: str, payload: dict = Body(...)):
    # Safety Check
    safety_result = control_plane.evaluate({"type": "tool_execution", "tool": tool_name, "payload": payload})
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    result = registry.run(tool_name, payload)
    if not result.success:
        return {"success": False, "error": result.error}
    return {"success": True, "output": result.output}

@app.get("/semantic/search")
async def semantic_search(q: str, top_k: int = 5):
    results = semantic.search(q, top_k=top_k)
    return {"results": results}

@app.get("/reasoning/influence/{entity}")
async def get_influence(entity: str):
    score = memory.calculate_influence_score(entity)
    return {"entity": entity, "influence_score": score}

@app.get("/reasoning/causal-chains/{entity}")
async def get_causal_chains(entity: str, max_depth: int = Query(4)):
    chains = memory.find_causal_chains(entity, max_depth=max_depth)
    return {"entity": entity, "causal_chains": chains}

@app.get("/reasoning/decisions/ranked")
async def get_ranked_decisions():
    decisions = memory.rank_decisions_by_impact()
    return {"ranked_decisions": decisions}

@app.get("/reasoning/path")
async def get_path(source: str, target: str, max_depth: int = Query(3)):
    path = memory.find_path(source, target, max_depth=max_depth)
    return {"source": source, "target": target, "path": path}

@app.post("/orchestrate")
async def orchestrate_task(task: str):
    # Safety Check
    safety_result = control_plane.evaluate({"type": "orchestration", "task": task})
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    result = orchestrator.run(task)
    return result
