CREATE TABLE ai_daily_usage (
    user_id BIGINT NOT NULL,
    usage_date DATE NOT NULL,
    request_count INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (user_id, usage_date)
);
