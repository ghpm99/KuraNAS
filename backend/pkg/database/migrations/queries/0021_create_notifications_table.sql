CREATE TABLE IF NOT EXISTS notifications (
    id SERIAL PRIMARY KEY,
    type VARCHAR(20) NOT NULL CHECK (type IN ('info','success','warning','error','system')),
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata JSON,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    group_key TEXT,
    group_count INTEGER NOT NULL DEFAULT 1,
    is_grouped BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_notifications_is_read_created ON notifications(is_read, created_at DESC);
CREATE INDEX idx_notifications_type_created ON notifications(type, created_at DESC);
CREATE INDEX idx_notifications_group_lookup ON notifications(group_key, type, created_at DESC) WHERE group_key IS NOT NULL;
