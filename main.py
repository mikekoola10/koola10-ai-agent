import asyncio
import os
import time
from typing import Optional
from fastapi import FastAPI, Body, Query, HTTPException
from memory_graph import MemoryGraph
from tools import registry
from semantic_memory import SemanticMemory
from agents import Orchestrator
from safety import ControlPlane
from business import BusinessLoop
from swarm import SwarmBus, SwarmNode
from swarm.cloud import CloudSwarmRegistry, NodeInfo
from swarm.router import TaskRouter
from production import ProductGenerator, ProductDeployer

app = FastAPI()
memory = MemoryGraph()
semantic = SemanticMemory()
orchestrator = Orchestrator()
control_plane = ControlPlane(threshold=0.6)
business = BusinessLoop()
swarm_bus = SwarmBus()
swarm_node = SwarmNode(swarm_bus)
product_generator = ProductGenerator()
product_deployer = ProductDeployer()

# Cross-Cloud Layer
cloud_registry = CloudSwarmRegistry(swarm_bus.client)
task_router = TaskRouter(cloud_registry)

NODE_ID = os.getenv("NODE_ID", "koola10-primary")
FLY_REGION = os.getenv("FLY_REGION", "local")
ENDPOINT = os.getenv("ENDPOINT", "http://localhost:8080")

@app.on_event("startup")
async def startup_event():
    # Register Node
    node_info = NodeInfo(
        node_id=NODE_ID,
        region=FLY_REGION,
        endpoint=ENDPOINT,
        last_seen=time.time()
    )
    cloud_registry.register(node_info)

    # Start Heartbeat
    asyncio.create_task(heartbeat_task())

async def heartbeat_task():
    while True:
        await asyncio.sleep(30)
        try:
            cloud_registry.heartbeat(NODE_ID)
        except Exception as e:
            print(f"Heartbeat failed: {e}")

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

@app.get("/swarm/cloud/nodes")
async def list_cloud_nodes(region: Optional[str] = None):
    if region:
        return {"nodes": cloud_registry.get_nodes_in_region(region)}
    return {"nodes": cloud_registry.get_all_healthy_nodes()}

@app.post("/swarm/cloud/route")
async def route_swarm_task(preferred_region: Optional[str] = None):
    node = task_router.select_optimal_node(preferred_region)
    if not node:
        raise HTTPException(status_code=503, detail="No healthy swarm nodes available")
    return {"selected_node": node}

@app.post("/production/generate")
async def generate_product(spec: dict = Body(...)):
    # Safety Check
    safety_result = control_plane.evaluate({"type": "product_generation", "spec": spec})
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    return product_generator.generate(spec)

@app.post("/production/deploy")
async def deploy_product(product_name: str):
    # Safety Check
    safety_result = control_plane.evaluate({"type": "product_deployment", "product_name": product_name})
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    product_dir = os.path.join(product_generator.products_path, product_name)
    if not os.path.exists(product_dir):
        raise HTTPException(status_code=404, detail="Product directory not found")

    return product_deployer.deploy(product_dir, product_name)

@app.post("/production/full")
async def full_production_cycle(spec: dict = Body(...)):
    # Safety Check
    safety_result = control_plane.evaluate({"type": "full_production", "spec": spec})
    if safety_result["decision"] == "BLOCK":
        raise HTTPException(status_code=403, detail=f"Action blocked by safety control plane. Risk score: {safety_result['risk_score']}")

    # Generate
    gen_result = product_generator.generate(spec)
    # Deploy
    deploy_result = product_deployer.deploy(gen_result["directory"], gen_result["product_name"])

    # Log to business loop
    business.log_action({"type": "full_production", "spec": spec}, {"gen": gen_result, "deploy": deploy_result})

    return {"generation": gen_result, "deployment": deploy_result}
