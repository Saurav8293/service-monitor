-- Create notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id          BIGSERIAL PRIMARY KEY,
    incident_id BIGINT NOT NULL,
    channel     TEXT NOT NULL DEFAULT 'log', -- log, slack, email
    message     TEXT NOT NULL,
    sent_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_notifications_incident ON notifications(incident_id);
