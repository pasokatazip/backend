from fastapi import FastAPI

app = FastAPI()


@app.get("/health")
def health():
    # Python動作確認用API
    return {"status": "ok"}


@app.post("/vectorize")
def vectorize():
    # ベクトル処理などを書く
    return {
        "message": "vectorize placeholder"
    }