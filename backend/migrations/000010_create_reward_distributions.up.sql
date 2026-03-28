CREATE TABLE reward_distributions (
    id             UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id    UUID      NOT NULL REFERENCES reward_campaigns(id) ON DELETE CASCADE,
    user_id        UUID      NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reward_type_id UUID      NOT NULL REFERENCES reward_types(id) ON DELETE CASCADE,
    status         TEXT      NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','delivered','failed')),
    delivered_at   TIMESTAMP,
    email_sent_at  TIMESTAMP
);
