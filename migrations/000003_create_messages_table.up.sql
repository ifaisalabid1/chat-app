CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL CHECK (
        length(content) > 0
        AND length(content) <= 4000
    ),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);