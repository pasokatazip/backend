import logging
from sentence_transformers import SentenceTransformer

logger = logging.getLogger(__name__)

logger.info("loading sentence-transformers model")
model = SentenceTransformer(
    "sentence-transformers/paraphrase-multilingual-MiniLM-L12-v2"
)
logger.info("model loaded")


def create_embedding(content: str) -> list[float]:
    logger.info("create_embedding start content_len=%d", len(content or ""))
    embedding = model.encode(
        content,
        normalize_embeddings=True,
    )
    logger.info("create_embedding done vector_len=%d", len(embedding))

    return embedding.tolist()


def to_pgvector(embedding: list[float]) -> str:
    return "[" + ",".join(map(str, embedding)) + "]"