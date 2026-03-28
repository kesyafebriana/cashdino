CREATE TABLE leaderboard_entries (
    id                  UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id        UUID        NOT NULL REFERENCES weekly_challenges(id) ON DELETE CASCADE,
    user_id             UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    weekly_gems         INT         NOT NULL DEFAULT 0,
    first_gem_earned_at TIMESTAMP,
    display_name        VARCHAR(20) NOT NULL,
    UNIQUE (challenge_id, user_id)
);

CREATE INDEX idx_leaderboard_entries_rank ON leaderboard_entries (challenge_id, weekly_gems DESC, first_gem_earned_at ASC);
