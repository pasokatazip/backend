from dataclasses import dataclass
import logging
import re

from sudachipy import dictionary
from sudachipy import tokenizer as sudachi_tokenizer

logger = logging.getLogger(__name__)

tokenizer = dictionary.Dictionary().create()
split_mode = sudachi_tokenizer.Tokenizer.SplitMode.C

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

STOP_VERBS = {
    "ある",
    "いる",
    "する",
    "なる",
    "できる",
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

    for token in tokenizer.tokenize(content, split_mode):
        part_of_speech = token.part_of_speech()
        if not part_of_speech or part_of_speech[0] != "名詞":
            continue
        if should_ignore_noun_part(part_of_speech):
            continue

        noun_text = token.surface().strip()
        normalized_noun = normalize_noun(token.normalized_form())
        if should_ignore_normalized_noun(normalized_noun) or normalized_noun in seen:
            continue

        seen.add(normalized_noun)
        nouns.append(
            ExtractedNoun(
                noun_text=noun_text,
                normalized_noun=normalized_noun,
            )
        )

    if nouns:
        logger.info("extract_nouns done noun_count=%d", len(nouns))
        return nouns

    verbs = extract_verbs_as_nouns(content)
    logger.info("extract_nouns fallback verb_count=%d", len(verbs))
    return verbs


def extract_verbs_as_nouns(content: str) -> list[ExtractedNoun]:
    verbs: list[ExtractedNoun] = []
    seen: set[str] = set()

    for token in tokenizer.tokenize(content, split_mode):
        part_of_speech = token.part_of_speech()
        if not part_of_speech or part_of_speech[0] != "動詞":
            continue

        verb_text = token.surface().strip()
        base_form = token.dictionary_form()
        normalized_verb = normalize_verb(base_form)
        if should_ignore_normalized_verb(normalized_verb) or normalized_verb in seen:
            continue

        seen.add(normalized_verb)
        verbs.append(
            ExtractedNoun(
                noun_text=verb_text,
                normalized_noun=normalized_verb,
            )
        )

    return verbs


def normalize_noun(noun: str) -> str:
    normalized = noun.strip().lower()
    normalized = re.sub(r"\s+", "", normalized)
    return normalized


def normalize_verb(verb: str) -> str:
    normalized = verb.strip().lower()
    normalized = re.sub(r"\s+", "", normalized)
    return normalized


def should_ignore_noun_part(part_of_speech: list[str]) -> bool:
    return any(detail in IGNORED_NOUN_DETAILS for detail in part_of_speech[1:])


def should_ignore_normalized_noun(normalized_noun: str) -> bool:
    if not normalized_noun:
        return True

    return normalized_noun in STOP_NORMALIZED_NOUNS


def should_ignore_normalized_verb(normalized_verb: str) -> bool:
    if not normalized_verb:
        return True

    return normalized_verb in STOP_VERBS
