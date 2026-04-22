from fastapi import FastAPI, Body, Query, HTTPException
from memory_graph import MemoryGraph
from tools import registry
from semantic_memory import SemanticMemory
from agents import Orchestrator
from safety import ControlPlane
from business import BusinessLoop
from swarm import SwarmBus, SwarmNode

app = FastAPI()
memory = MemoryGraph()
semantic = SemanticMemory()
orchestrator = Orchestrator()
control_plane = ControlPlane(threshold=0.6)
business = BusinessLoop()
swarm_bus = SwarmBus()
swarm_node = SwarmNode(swarm_bus)

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
    action_info = {"type": "tool_execution", "tool": tool_name, "payload": payload}

    # Safety Check
    safety_result = control_plane.evaluate(action_info)
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    result = registry.run(tool_name, payload)

    # Log Action in Business Loop
    action_id = business.log_action(action_info, result.__dict__ if hasattr(result, "__dict__") else result)

    if not result.success:
        return {"success": False, "error": result.error, "action_id": action_id}
    return {"success": True, "output": result.output, "action_id": action_id}

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

@app.post("/business/outcome")
async def report_outcome(action_id: str, outcome: dict = Body(...)):
    success = business.log_outcome(action_id, outcome)
    if not success:
        raise HTTPException(status_code=404, detail="Action ID not found")
    return {"status": "outcome logged"}

@app.get("/business/metrics")
async def get_metrics():
    return business.get_feedback_signal()

@app.post("/swarm/broadcast")
async def broadcast_task(task: str):
    msg = swarm_node.broadcast_task(task)
    return {"status": "broadcasted", "message": msg}

@app.post("/swarm/result")
async def send_swarm_result(result: dict = Body(...)):
    msg = swarm_node.send_result(result)
    return {"status": "result_sent", "message": msg}

@app.get("/swarm/listen")
async def listen_swarm(limit: int = 10):
    messages = swarm_bus.get_recent_messages("results", limit=limit)
    return {"recent_results": messages}
