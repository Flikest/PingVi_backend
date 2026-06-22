CREATE KEYSPACE IF NOT EXISTS messenger_keyspace
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};

USE messenger_keyspace;

-- ============================================
-- 2. КАНАЛЫ (сервера как в Discord)
-- ============================================
CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY,
    name TEXT,
    description TEXT,
    owner_id UUID,
    icon_urls LIST<TEXT>,
    is_public BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS directory (
    id UUID NOT NULL,
    channel_id UUID NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP
);

-- ============================================
-- 3. ТОПИКИ (текстовые/голосовые каналы внутри сервера)
-- ============================================
CREATE TABLE IF NOT EXISTS topics (
    id UUID PRIMARY KEY,
    channel_id UUID,                -- В каком канале (сервере) находится топик
    directory_id UUID,
    name TEXT,                      -- Название топика (например, "general", "offtopic")
    topic_type TEXT,                -- 'text', 'voice', 'announcement'
    position INT,                   -- Позиция в списке (для сортировки)
    is_private BOOLEAN,             -- Приватный канал (видят только участники с ролью)
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- ============================================
-- 4. РОЛИ В КАНАЛЕ (привилегии на весь канал/сервер)
-- ============================================
CREATE TABLE IF NOT EXISTS channel_roles (
    id UUID PRIMARY KEY,
    channel_id UUID,                
    name TEXT,                      -- '@everyone', '@Admin', '@Moderator', '@Member'
    color INT,                      -- Цвет роли (HEX)
    permissions varchar(7) is_default '0000000',             -- БИТОВАЯ МАСКА (разрешения на весь канал) 000000 [1] - взаимодействие с пользователями(удалить\добавить) [2] - выдават\забирать роли [3] - взаимодействие с каналом (обновить\удалить) [4] - писать сообщения в канал  [5] - создавать\удалять\обновлять роли, а так же содовать вложенности [6] - взаимодействие с топиками [7] - учавствовать в топиках
    is_mentionable BOOLEAN,
    is_default BOOLEAN,             -- true для роли @everyone
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- ============================================
-- 5. УЧАСТНИКИ КАНАЛА (с их ролями)
-- Привилегии распространяются на ВЕСЬ канал и все топики внутри него
-- ============================================
CREATE TABLE IF NOT EXISTS channel_members (
    channel_id UUID,
    user_id UUID,
    is_muted BOOLEAN,
    role_id UUID,                   -- Основная роль (ссылка на channel_roles)
    nickname TEXT,                  -- Ник в рамках канала (сервера)
    joined_at TIMESTAMP,
    PRIMARY KEY (channel_id, user_id)
);

-- ============================================
-- 7. ЛИЧНЫЕ ЧАТЫ (вне каналов, на 2 пользователя)
-- ============================================
CREATE TABLE IF NOT EXISTS personal_chats (
    id UUID PRIMARY KEY,
    user1_id UUID,
    user2_id UUID,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
);

-- ============================================
-- 8. ГРУППЫ (большие чаты без каналов и привилегий)
-- ============================================
CREATE TABLE IF NOT EXISTS groups (
    id UUID PRIMARY KEY,
    name TEXT,
    avatar_links LIST<TEXT>,
    owner_id UUID,
    is_public BOOLEAN,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
);

-- ============================================
-- 9. УЧАСТНИКИ ГРУПП
-- ============================================
CREATE TABLE IF NOT EXISTS group_members (
    group_id UUID,
    user_id UUID,
    is_muted BOOLEAN NOT NULL,
    is_admin BOOLEAN NOT NULL,
    joined_at TIMESTAMP NOT NULL,
    PRIMARY KEY (group_id, user_id)
);

-- ============================================
-- 10. СООБЩЕНИЯ (универсальная)
-- chat_id может быть: topic_id, personal_chat_id, group_id
-- ============================================
CREATE TABLE IF NOT EXISTS messages (
    id UUID,
    chat_id UUID,
    sender_id UUID,
    message TEXT,
    message_type TEXT,              -- 'text', 'image', 'video', 'file', 'sticker'
    is_edited BOOLEAN,
    reply_to_id TIMEUUID,
    attachments LIST<TEXT>,
    reactions MAP<TEXT, LIST<UUID>>,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    PRIMARY KEY ((chat_id, date_bucket), message_id)
) WITH CLUSTERING ORDER BY (id ASC);

-- ============================================
-- 12. ПРОЧИТАННЫЕ СООБЩЕНИЯ
-- ============================================
CREATE TABLE IF NOT EXISTS user_read_positions (
    user_id UUID,
    chat_id UUID,
    last_read_message_id TIMEUUID,
    last_read_time TIMESTAMP,
    PRIMARY KEY (user_id, chat_id)
);

-- ============================================
-- 13. ПРИГЛАСИТЕЛЬНЫЕ ССЫЛКИ (для каналов и групп)
-- ============================================
CREATE TABLE IF NOT EXISTS invitations (
    id UUID PRIMARY KEY,
    target_type TEXT,               -- 'channel', 'group'
    target_id UUID,
    invite_code TEXT UNIQUE,
    created_by UUID,
    max_uses INT,
    used_count INT,
    expires_at TIMESTAMP,
    created_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_invite_code ON invitations (invite_code);
