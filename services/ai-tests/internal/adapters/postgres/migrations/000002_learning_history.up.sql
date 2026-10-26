CREATE INDEX idx_adaptive_sessions_learning_history
    ON adaptive_sessions (user_id, course_id, topic_page_id, completed_at DESC)
    WHERE status = 'completed';
