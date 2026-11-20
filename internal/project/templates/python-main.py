from fastapi import FastAPI
app = FastAPI(title="DevForge Python Service")

@app.get("/health")
def health():
    return {"status": "ok"}
