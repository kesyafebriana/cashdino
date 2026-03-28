CREATE TABLE daily_checkins (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    date              DATE          NOT NULL UNIQUE,
    base_gems         INT           NOT NULL,
    streak_multiplier DECIMAL(3,2)  NOT NULL DEFAULT 1.00,
    is_active         BOOLEAN       NOT NULL DEFAULT true
);
