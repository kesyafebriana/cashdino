CREATE TABLE reward_campaigns (
    id                          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id                UUID         NOT NULL REFERENCES weekly_challenges(id) ON DELETE CASCADE,
    name                        VARCHAR(100) NOT NULL,
    banner_image                VARCHAR(500),
    rules                       JSONB        NOT NULL DEFAULT '[]',
    non_gem_claim_email_subject VARCHAR(255),
    non_gem_claim_email_body    TEXT,
    status                      TEXT         NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','scheduled','active','completed'))
);
