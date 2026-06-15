from sentence_transformers import SentenceTransformer

model = SentenceTransformer(
    "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
)


def create_embedding(content: str) -> list[float]:
    embedding = model.encode(
        content,
        normalize_embeddings=True,
    )

    return embedding.tolist()


def to_pgvector(embedding: list[float]) -> str:
    return "[" + ",".join(map(str, embedding)) + "]"