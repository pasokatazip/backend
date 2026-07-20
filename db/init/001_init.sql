--pg-vecter
CREATE EXTENSION IF NOT EXISTS vector;

--users
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    subsc BOOLEAN NOT NULL DEFAULT FALSE,
    fincode_customer_id VARCHAR(255) UNIQUE,
    fincode_subscription_id VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- notifications
CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id),
    is_all_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_yoyo_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_report_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    is_message_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    subscription JSONB NOT NULL
);

-- pets
CREATE TABLE pets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    color VARCHAR(7) NOT NULL DEFAULT '#FFC1CA',
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    user_id UUID NOT NULL REFERENCES users(id),
    energy DECIMAL(7, 4) DEFAULT 50,
    curiosity DECIMAL(7, 4) DEFAULT 50,
    sociality DECIMAL(7, 4) DEFAULT 50,
    routine DECIMAL(7, 4) DEFAULT 50,
    current_group_master_id INTEGER,
    current_stage_id INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_pets_status CHECK (
        status IN ('active', 'departed', 'lost', 'archived')
    )
);

CREATE INDEX IF NOT EXISTS idx_pets_status ON pets(status);

-- USER_ACTIVE_PETS(
CREATE TABLE user_active_pets (
    user_id UUID NOT NULL REFERENCES users(id),
    pet_id UUID NOT NULL REFERENCES pets(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, pet_id)
);

-- PET_EXPERIENCES
CREATE TABLE pet_experiences (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL UNIQUE REFERENCES pets(id),
    total_experience BIGINT NOT NULL DEFAULT 0,
    feed_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- PET_EXPERIENCE_EVENTS
CREATE TABLE pet_experience_events (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    source_type VARCHAR(50) NOT NULL,
    source_id UUID,
    amount INT NOT NULL,
    capped_amount INT NOT NULL DEFAULT 0,
    experience_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pet_experience_events_pet_date ON pet_experience_events(pet_id, experience_date);

CREATE INDEX IF NOT EXISTS idx_pet_experience_events_source_type ON pet_experience_events(source_type);

-- EXPERIENCE_CAPS
CREATE TABLE experience_caps (
    id UUID PRIMARY KEY,
    cap_type VARCHAR(50) NOT NULL,
    max_experience INT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_experience_caps_type_active ON experience_caps(cap_type, active);

-- EVOLUTION_STAGES
CREATE TABLE evolution_stages (
    id INT PRIMARY KEY,
    stage_key VARCHAR(100) NOT NULL UNIQUE,
    stage_no INT NOT NULL,
    name VARCHAR(100) NOT NULL,
    branch_key VARCHAR(100),
    image_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_evolution_stages_stage_no ON evolution_stages(stage_no);

CREATE INDEX IF NOT EXISTS idx_evolution_stages_branch_key ON evolution_stages(branch_key);

ALTER TABLE
    pets
ADD
    CONSTRAINT fk_pets_current_stage FOREIGN KEY (current_stage_id) REFERENCES evolution_stages(id);

-- EVOLUTION_RULES
CREATE TABLE evolution_rules (
    id SERIAL PRIMARY KEY,
    from_stage_id INT NOT NULL REFERENCES evolution_stages(id),
    to_stage_id INT NOT NULL REFERENCES evolution_stages(id),
    required_experience BIGINT NOT NULL DEFAULT 0,
    required_days_since_last_evolution INT NOT NULL DEFAULT 0,
    required_feed_count INT NOT NULL DEFAULT 0,
    appearance_part VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (from_stage_id, to_stage_id)
);

CREATE INDEX IF NOT EXISTS idx_evolution_rules_from_stage ON evolution_rules(from_stage_id);

-- PET_EVOLUTIONS
CREATE TABLE pet_evolutions (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    stage_id INT NOT NULL REFERENCES evolution_stages(id),
    evolution_rule_id INT REFERENCES evolution_rules(id),
    primary_status VARCHAR(50),
    evolved_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pet_evolutions_pet_id ON pet_evolutions(pet_id);

CREATE INDEX IF NOT EXISTS idx_pet_evolutions_stage_id ON pet_evolutions(stage_id);

--posts
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    content VARCHAR(255) NOT NULL,
    content_embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--group_masters
CREATE TABLE IF NOT EXISTS group_masters (
    id SERIAL PRIMARY KEY,
    group_key VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    display_name_embedding vector(384),
    category VARCHAR(100),
    min_pet_count INTEGER NOT NULL DEFAULT 0,
    energy_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    curiosity_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    sociality_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    routine_delta DECIMAL(5, 4) NOT NULL DEFAULT 0,
    morning_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    afternoon_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    night_weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_masters_active ON group_masters(active);

CREATE INDEX IF NOT EXISTS idx_group_masters_category ON group_masters(category);

ALTER TABLE
    pets
ADD
    CONSTRAINT fk_pets_current_group_master FOREIGN KEY (current_group_master_id) REFERENCES group_masters(id);

--group_keywords
CREATE TABLE IF NOT EXISTS group_keywords (
    id SERIAL PRIMARY KEY,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    keyword VARCHAR(100) NOT NULL,
    normalized_keyword VARCHAR(100) NOT NULL,
    weight DECIMAL(5, 2) NOT NULL DEFAULT 1.00,
    match_type VARCHAR(50) NOT NULL DEFAULT 'exact_or_partial',
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_group_keywords_lookup ON group_keywords (normalized_keyword, active);

CREATE UNIQUE INDEX IF NOT EXISTS uq_group_keywords_group_keyword_match ON group_keywords (group_master_id, normalized_keyword, match_type);

--extracted_nouns
CREATE TABLE IF NOT EXISTS extracted_nouns (
    id SERIAL PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    noun_text VARCHAR(255) NOT NULL,
    normalized_noun VARCHAR(255) NOT NULL,
    noun_embedding VECTOR(384),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_extracted_nouns_post_id ON extracted_nouns(post_id);

CREATE INDEX IF NOT EXISTS idx_extracted_nouns_normalized_noun ON extracted_nouns(normalized_noun);

--noun_group_matches
CREATE TABLE IF NOT EXISTS noun_group_matches (
    id SERIAL PRIMARY KEY,
    extracted_noun_id INTEGER NOT NULL REFERENCES extracted_nouns(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    keyword_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    vector_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    keyword_weight DECIMAL(5, 2) NOT NULL DEFAULT 0,
    match_score DECIMAL(8, 5) NOT NULL DEFAULT 0,
    match_reason VARCHAR(255),
    selected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

--report
CREATE TABLE reports (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id),
    hour_slot INTEGER NOT NULL CHECK (
        hour_slot BETWEEN 0
        AND 23
    ),
    gossip VARCHAR(255),
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    previous_group_master_id INTEGER,
    moved BOOLEAN NOT NULL DEFAULT FALSE,
    behavior_type VARCHAR(50) NOT NULL,
    behavior_label VARCHAR(255) NOT NULL,
    interaction_count INTEGER NOT NULL DEFAULT 0,
    energy_delta INTEGER NOT NULL DEFAULT 0,
    curiosity_delta INTEGER NOT NULL DEFAULT 0,
    sociality_delta INTEGER NOT NULL DEFAULT 0,
    routine_delta INTEGER NOT NULL DEFAULT 0,
    reason_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_extracted_noun_id ON noun_group_matches(extracted_noun_id);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_group_master_id ON noun_group_matches(group_master_id);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_selected ON noun_group_matches(selected);

CREATE INDEX IF NOT EXISTS idx_noun_group_matches_match_score ON noun_group_matches(match_score);

-- pet_group_interests
-- 投稿ごとの一致履歴ではなく、ペットが関心を持つ群れの累積スコアを保持
CREATE TABLE IF NOT EXISTS pet_group_interests (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    interest_score DECIMAL(12, 5) NOT NULL DEFAULT 0,
    last_matched_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_pet_group_interests_pet_group UNIQUE (pet_id, group_master_id),
    CONSTRAINT chk_pet_group_interests_score CHECK (interest_score >= 0)
);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_pet_score ON pet_group_interests(pet_id, interest_score DESC);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_group_master_id ON pet_group_interests(group_master_id);

CREATE INDEX IF NOT EXISTS idx_pet_group_interests_last_matched_at ON pet_group_interests(last_matched_at);

-- pet_group_joins
CREATE TABLE IF NOT EXISTS pet_group_joins (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMPTZ,
    move_reason VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pet_group_joins_pet_active ON pet_group_joins(pet_id, left_at);

CREATE INDEX IF NOT EXISTS idx_pet_group_joins_group_active ON pet_group_joins(group_master_id, left_at);

CREATE INDEX IF NOT EXISTS idx_pet_group_joins_joined_at ON pet_group_joins(joined_at);

-- pet_hourly_logs
CREATE TABLE IF NOT EXISTS pet_hourly_logs (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    pet_group_join_id UUID REFERENCES pet_group_joins(id),
    simulated_at TIMESTAMPTZ NOT NULL,
    stayed BOOLEAN NOT NULL DEFAULT TRUE,
    move_probability DECIMAL(6, 5) NOT NULL DEFAULT 0,
    boredom DECIMAL(6, 5) NOT NULL DEFAULT 0,
    rest_need DECIMAL(6, 5) NOT NULL DEFAULT 0,
    current_group_fit DECIMAL(6, 5) NOT NULL DEFAULT 0,
    attachment_to_current_group DECIMAL(6, 5) NOT NULL DEFAULT 0,
    recent_move_penalty DECIMAL(6, 5) NOT NULL DEFAULT 0,
    energy_delta_applied DECIMAL(5, 4) NOT NULL DEFAULT 0,
    curiosity_delta_applied DECIMAL(5, 4) NOT NULL DEFAULT 0,
    sociality_delta_applied DECIMAL(5, 4) NOT NULL DEFAULT 0,
    routine_delta_applied DECIMAL(5, 4) NOT NULL DEFAULT 0,
    interaction_count INTEGER NOT NULL DEFAULT 0,
    ambient_event VARCHAR(100),
    report_material TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (pet_id, simulated_at)
);

CREATE INDEX IF NOT EXISTS idx_pet_hourly_logs_group_master_id ON pet_hourly_logs(group_master_id);

CREATE INDEX IF NOT EXISTS idx_pet_hourly_logs_simulated_at ON pet_hourly_logs(simulated_at);

-- pet_interest_propagations
-- 同じ時間・同じ群れにいた別ペットから伝わった興味の履歴
-- 投稿本文・抽出名詞は保存せず、群れとスコアだけを保持
CREATE TABLE IF NOT EXISTS pet_interest_propagations (
    id UUID PRIMARY KEY,
    recipient_pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    source_pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    source_pet_hourly_log_id UUID NOT NULL REFERENCES pet_hourly_logs(id) ON DELETE CASCADE,
    via_group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    propagated_group_master_id INTEGER NOT NULL REFERENCES group_masters(id),
    amount DECIMAL(8, 5) NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_pet_interest_propagations_different_pets CHECK (recipient_pet_id <> source_pet_id),
    CONSTRAINT chk_pet_interest_propagations_amount CHECK (amount > 0),
    CONSTRAINT uq_pet_interest_propagations_source UNIQUE (
        recipient_pet_id,
        source_pet_hourly_log_id,
        propagated_group_master_id
    )
);

CREATE INDEX IF NOT EXISTS idx_pet_interest_propagations_recipient_occurred_at ON pet_interest_propagations(recipient_pet_id, occurred_at DESC);

CREATE INDEX IF NOT EXISTS idx_pet_interest_propagations_source_hourly_log_id ON pet_interest_propagations(source_pet_hourly_log_id);

-- souvenir_masters
CREATE TABLE IF NOT EXISTS souvenir_masters (
    id SERIAL PRIMARY KEY,
    souvenir_key VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    image_url TEXT,
    group_master_id INTEGER NOT NULL REFERENCES group_masters(id) ON DELETE CASCADE,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_souvenir_masters_group_master_id UNIQUE (group_master_id)
);

CREATE INDEX IF NOT EXISTS idx_souvenir_masters_active ON souvenir_masters(active);

-- pet_souvenirs
CREATE TABLE IF NOT EXISTS pet_souvenirs (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    souvenir_master_id INTEGER NOT NULL REFERENCES souvenir_masters(id),
    pet_hourly_log_id UUID REFERENCES pet_hourly_logs(id),
    report_id UUID REFERENCES reports(id),
    found_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    found_on DATE NOT NULL,
    source_group_master_id INTEGER REFERENCES group_masters(id),
    note TEXT,
    reported_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_pet_id ON pet_souvenirs(pet_id);

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_souvenir_master_id ON pet_souvenirs(souvenir_master_id);

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_hourly_log_id ON pet_souvenirs(pet_hourly_log_id);

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_report_id ON pet_souvenirs(report_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_pet_souvenirs_report_id ON pet_souvenirs(report_id)
WHERE
    report_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_pet_found_on ON pet_souvenirs(pet_id, found_on);

CREATE INDEX IF NOT EXISTS idx_pet_souvenirs_reported_at ON pet_souvenirs(reported_at);

-- pet_departure_rules
CREATE TABLE IF NOT EXISTS pet_departure_rules (
    id SERIAL PRIMARY KEY,
    rule_key VARCHAR(100) NOT NULL UNIQUE,
    min_age_days INTEGER NOT NULL DEFAULT 30,
    required_stage_id INTEGER NOT NULL DEFAULT 2 REFERENCES evolution_stages(id),
    grace_days_min INTEGER NOT NULL DEFAULT 7,
    grace_days_max INTEGER NOT NULL DEFAULT 7,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_pet_departure_rules_min_age CHECK (min_age_days >= 30),
    CONSTRAINT chk_pet_departure_rules_grace_days CHECK (
        grace_days_min >= 0
        AND grace_days_max >= grace_days_min
    )
);

CREATE INDEX IF NOT EXISTS idx_pet_departure_rules_active ON pet_departure_rules(active);

CREATE INDEX IF NOT EXISTS idx_pet_departure_rules_required_stage_id ON pet_departure_rules(required_stage_id);

-- pet_departures
CREATE TABLE IF NOT EXISTS pet_departures (
    id UUID PRIMARY KEY,
    pet_id UUID NOT NULL UNIQUE REFERENCES pets(id),
    user_id UUID NOT NULL REFERENCES users(id),
    pet_departure_rule_id INTEGER REFERENCES pet_departure_rules(id),
    eligible_at TIMESTAMPTZ,
    scheduled_departure_at TIMESTAMPTZ,
    departed_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'waiting',
    blocked_reason VARCHAR(255),
    farewell_report_id UUID REFERENCES reports(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_pet_departures_status CHECK (
        status IN (
            'waiting',
            'eligible',
            'scheduled',
            'departed',
            'blocked'
        )
    ),
    CONSTRAINT chk_pet_departures_schedule CHECK (
        scheduled_departure_at IS NULL
        OR eligible_at IS NULL
        OR scheduled_departure_at >= eligible_at
    ),
    CONSTRAINT chk_pet_departures_departed CHECK (
        departed_at IS NULL
        OR scheduled_departure_at IS NULL
        OR departed_at >= scheduled_departure_at
    )
);

CREATE INDEX IF NOT EXISTS idx_pet_departures_user_id ON pet_departures(user_id);

CREATE INDEX IF NOT EXISTS idx_pet_departures_status ON pet_departures(status);

CREATE INDEX IF NOT EXISTS idx_pet_departures_eligible_at ON pet_departures(eligible_at);

CREATE INDEX IF NOT EXISTS idx_pet_departures_scheduled_departure_at ON pet_departures(scheduled_departure_at);

CREATE INDEX IF NOT EXISTS idx_pet_departures_departed_at ON pet_departures(departed_at);
