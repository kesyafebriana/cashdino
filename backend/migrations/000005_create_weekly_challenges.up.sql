CREATE TABLE weekly_challenges (
    id         UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    start_time TIMESTAMP NOT NULL,
    end_time   TIMESTAMP NOT NULL,
    status     TEXT      NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled','active','completed'))
);
