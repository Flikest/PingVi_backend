CREATE TABLE IF NOT EXISTS bots (
    id UUID PRIMARY KEY,
    name TEXT not null,
    description TEXT,
    avatar TEXT[],
    creator_id UUID NOT NULL,
    token TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_bots_id ON bots(id);
CREATE INDEX IF NOT EXISTS idx_bots_creator_id ON bots(creator_id);

CREATE TABLE IF NOT EXISTS bot_commands (
    id UUID PRIMARY KEY,
    bot_id UUID NOT NULL,
    command TEXT NOT NULL,
    description TEXT NOT NULL,

    CONSTRAINT fk_bot_comands_bots
        FOREIGN KEY (bot_id)
        REFERENCES bots(id)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_bot_commands_id ON bot_commands(id)
CREATE INDEX IF NOT EXISTS idx_bot_commands_bot_id ON bot_commands(bot_id)

