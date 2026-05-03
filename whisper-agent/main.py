import os
import whisper
import torch
from fastapi import FastAPI, UploadFile, File
import uvicorn
import shutil
import tempfile

app = FastAPI()

# Load model
model_size = os.getenv("WHISPER_MODEL", "small")
device = "cuda" if torch.cuda.is_available() else "cpu"
print(f"Loading Whisper model '{model_size}' on {device}...")
model = whisper.load_model(model_size, device=device)

@app.get("/health")
async def health():
    return {"status": "ok", "model": model_size, "device": device}

@app.post("/transcribe")
async def transcribe(file: UploadFile = File(...)):
    with tempfile.NamedTemporaryFile(delete=False, suffix=".wav") as tmp:
        shutil.copyfileobj(file.file, tmp)
        tmp_path = tmp.name

    try:
        result = model.transcribe(tmp_path)
        return {"text": result["text"]}
    finally:
        if os.path.exists(tmp_path):
            os.remove(tmp_path)

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8080)
