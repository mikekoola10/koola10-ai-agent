import os
from fastapi import FastAPI
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import numpy as np

app = FastAPI()

# Load model on startup
model_name = 'all-MiniLM-L6-v2'
model = SentenceTransformer(model_name)

class EmbeddingRequest(BaseModel):
    text: str

class SearchRequest(BaseModel):
    query: str
    embeddings: list # List of objects { "ref_id": "...", "vector": [...] }
    top_k: int = 5

@app.get("/health")
async def health():
    return {"status": "ok", "model": model_name}

@app.post("/generate")
async def generate(req: EmbeddingRequest):
    vector = model.encode(req.text).tolist()
    return {"vector": vector}

@app.post("/search")
async def search(req: SearchRequest):
    query_vec = model.encode(req.query)

    results = []
    for item in req.embeddings:
        target_vec = np.array(item["vector"])
        # Cosine similarity
        score = np.dot(query_vec, target_vec) / (np.linalg.norm(query_vec) * np.linalg.norm(target_vec))
        results.append({
            "ref_id": item["ref_id"],
            "score": float(score)
        })

    # Sort and return top_k
    results.sort(key=lambda x: x["score"], reverse=True)
    return results[:req.top_k]

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
