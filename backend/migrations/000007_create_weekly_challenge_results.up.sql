CREATE TABLE weekly_challenge_results (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID        NOT NULL REFERENCES weekly_challenges(id) ON DELETE CASCADE,
    user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    final_rank   INT         NOT NULL,
    final_gems   INT         NOT NULL,
    display_name VARCHAR(20) NOT NULL
);
