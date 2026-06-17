from dataclasses import dataclass
import logging
import re

from janome.tokenizer import Tokenizer

logger = logging.getLogger(__name__)

tokenizer = Tokenizer()


@dataclass(frozen=True)
class ExtractedNoun:
    noun_text: str
    normalized_noun: str


def extract_nouns(content: str) -> list[ExtractedNoun]:
    logger.info("extract_nouns start content_len=%d", len(content or ""))

    if not content:
        return []

    nouns: list[ExtractedNoun] = []
    seen: set[str] = set()

    for token in tokenizer.tokenize(content):
        part_of_speech = token.part_of_speech.split(",")
        if not part_of_speech or part_of_speech[0] != "名詞":
            continue

        noun_text = token.surface.strip()
        normalized_noun = normalize_noun(noun_text)
        if not normalized_noun or normalized_noun in seen:
            continue

        seen.add(normalized_noun)
        nouns.append(
            ExtractedNoun(
                noun_text=noun_text,
                normalized_noun=normalized_noun,
            )
        )

    logger.info("extract_nouns done noun_count=%d", len(nouns))
    return nouns


def normalize_noun(noun: str) -> str:
    normalized = noun.strip().lower()
    normalized = re.sub(r"\s+", "", normalized)
    return normalized
