from dataclasses import dataclass
import logging
import re

from janome.tokenizer import Tokenizer

logger = logging.getLogger(__name__)

tokenizer = Tokenizer()

STOP_NORMALIZED_NOUNS = {
    "今日",
    "昨日",
    "明日",
    "明後日",
    "一昨日",
    "今",
    "朝",
    "昼",
    "夜",
    "午前",
    "午後",
    "こと",
    "もの",
    "ため",
    "ところ",
    "感じ",
    "よう",
}

IGNORED_NOUN_DETAILS = {
    "非自立",
    "代名詞",
    "副詞可能",
}


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
        if should_ignore_noun_part(part_of_speech):
            continue

        noun_text = token.surface.strip()
        normalized_noun = normalize_noun(noun_text)
        if should_ignore_normalized_noun(normalized_noun) or normalized_noun in seen:
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


def should_ignore_noun_part(part_of_speech: list[str]) -> bool:
    return any(detail in IGNORED_NOUN_DETAILS for detail in part_of_speech[1:])


def should_ignore_normalized_noun(normalized_noun: str) -> bool:
    if not normalized_noun:
        return True

    return normalized_noun in STOP_NORMALIZED_NOUNS
