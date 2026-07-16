CREATE TABLE IF NOT EXISTS channels (
    id uuid PRIMARY KEY,
    name text,
    description text,
    owner_id uuid,
    icon_urls list<text>,
    is_public boolean,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS directory (
    id uuid PRIMARY KEY,
    channel_id uuid,
    name text,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS topics (
    id uuid PRIMARY KEY,
    channel_id uuid,
    directory_id uuid,
    name text,
    topic_type text,
    position int,
    is_private boolean,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS channel_roles (
    id uuid PRIMARY KEY,
    channel_id uuid,                
    name text,
    color int,
    permissions text,   -- БИТОВАЯ МАСКА (разрешения на весь канал) 000000 [1] - взаимодействие с пользователями(удалить\добавить) [2] - выдават\забирать роли [3] - взаимодействие с каналом (обновить\удалить) [4] - писать сообщения в канал  [5] - создавать\удалять\обновлять роли, а так же содовать вложенности [6] - взаимодействие с топиками [7] - учавствовать в топиках
    is_mentionable boolean,
    is_default boolean,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS channel_members (
    channel_id uuid,
    user_id uuid,
    is_muted boolean,
    role_id uuid,
    nickname text,
    joined_at timestamp,
    PRIMARY KEY (channel_id, user_id)
);

CREATE TABLE IF NOT EXISTS personal_chats (
    id uuid PRIMARY KEY,
    user1_id uuid,
    user2_id uuid,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS groups (
    id uuid PRIMARY KEY,
    name text,
    avatar_links list<text>,
    owner_id uuid,
    is_public boolean,
    created_at timestamp,
    updated_at timestamp
);

CREATE TABLE IF NOT EXISTS group_members (
    group_id uuid,
    user_id uuid,
    is_muted boolean,
    is_admin boolean,
    joined_at timestamp,
    PRIMARY KEY (group_id, user_id)
);


CREATE TABLE IF NOT EXISTS messages (
    id BIGINT,
    chat_id uuid,
    backet_id BIGINT,
    sender_id uuid,
    message text,
    reply_to_id BIGINT,
    created_at timestamp,
    updated_at timestamp,
    PRIMARY KEY ((chat_id, backet_id), id)
) WITH CLUSTERING ORDER BY (id ASC);

CREATE TABLE IF NOT EXISTS reactions (
    chat_id UUID,
    user_id UUID,
    message_id UUID,
    reaction TEXT,
    sended_at timestamp,
    PRIMARY KEY((chat_id, user_id))
);

CREATE TABLE IF NOT EXISTS user_read_positions (
    user_id uuid,
    chat_id uuid,
    last_read_message_id timeuuid,
    last_read_time timestamp,
    PRIMARY KEY (user_id, chat_id)
);

CREATE TABLE IF NOT EXISTS invitation_links (
    id uuid PRIMARY KEY,
    target_type text,
    target_id uuid,
    invite_code text, 
    created_by uuid,
    max_uses int,
    used_count int,
    expires_at timestamp,
    created_at timestamp
);

CREATE INDEX IF NOT EXISTS idx_invite_code ON invitation_links (invite_code);

