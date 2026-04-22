import json
import os
import numpy as np
from sentence_transformers import SentenceTransformer

class SemanticMemory:
    def __init__(self):
        self.storage_path = os.getenv("METACLAW_STORAGE_PATH", "/data")
        self.file_path = os.path.join(self.storage_path, "semantic_index.json")
        self.model = SentenceTransformer('all-MiniLM-L6-v2')
        self.data = self._load_data()

    def _load_data(self):
        if os.path.exists(self.file_path):
            with open(self.file_path, 'r') as f:
                loaded = json.load(f)
                # Convert list back to numpy array for embeddings
                for item in loaded:
                    item["embedding"] = np.array(item["embedding"])
                return loaded
        return []

    def _save_data(self):
        # Ensure directory exists
        os.makedirs(os.path.dirname(self.file_path), exist_ok=True)
        # Convert numpy arrays to lists for JSON serialization
        to_save = []
        for item in self.data:
            save_item = item.copy()
            save_item["embedding"] = item["embedding"].tolist()
            to_save.append(save_item)
        with open(self.file_path, 'w') as f:
            json.dump(to_save, f)

    def add_text(self, text, ref_id, meta=None):
        embedding = self.model.encode(text)
        self.data.append({
            "text": text,
            "ref_id": ref_id,
            "meta": meta or {},
            "embedding": embedding
        })
        self._save_data()

    def search(self, query, top_k=5):
        if not self.data:
            return []

        query_embedding = self.model.encode(query)

        results = []
        for item in self.data:
            # Cosine similarity
            sim = np.dot(query_embedding, item["embedding"]) / (
                np.linalg.norm(query_embedding) * np.linalg.norm(item["embedding"])
            )
            results.append({
                "text": item["text"],
                "ref_id": item["ref_id"],
                "meta": item["meta"],
                "score": float(sim)
            })

        # Sort by score descending
        results.sort(key=lambda x: x["score"], reverse=True)
        return results[:top_k]
