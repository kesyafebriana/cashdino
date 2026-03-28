CREATE TABLE user_daily_checkins (
    id             UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    checkin_id     UUID      NOT NULL REFERENCES daily_checkins(id) ON DELETE CASCADE,
    gems_earned    INT       NOT NULL,
    current_streak INT       NOT NULL DEFAULT 1,
    checked_in_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, checkin_id)
);
