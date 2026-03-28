CREATE TABLE gem_history (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source     TEXT         NOT NULL CHECK (source IN ('gameplay','daily_checkin','survey','referral','boost','reward','payout')),
    amount     INT          NOT NULL,
    game_name  VARCHAR(100),
    created_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gem_history_user_created ON gem_history (user_id, created_at DESC);
