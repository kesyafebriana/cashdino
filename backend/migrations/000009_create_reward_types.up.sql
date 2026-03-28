CREATE TABLE reward_types (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id UUID         NOT NULL REFERENCES reward_campaigns(id) ON DELETE CASCADE,
    name        VARCHAR(100) NOT NULL,
    type        TEXT         NOT NULL CHECK (type IN ('gems','gift_card','cash','other')),
    value       DECIMAL      NOT NULL,
    image       VARCHAR(500),
    stock       INT          NOT NULL DEFAULT 0
);
