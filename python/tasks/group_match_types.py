from dataclasses import dataclass


@dataclass(frozen=True)
class GroupMatchCandidate:
    group_master_id: int
    keyword_score: float
    vector_score: float
    keyword_weight: float
    match_score: float
    match_reason: str
    selectable: bool = True
