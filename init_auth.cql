CREATE USER IF NOT EXISTS pingvi WITH PASSWORD 'PJfoiu34298fn290djoshdf3hf2uhf021djpoj3098rh49uha9uh' SUPERUSER;
CREATE KEYSPACE IF NOT EXISTS messenger_keyspace WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
CREATE USER IF NOT EXISTS messenger_app WITH PASSWORD 'AppSecurePassword123';
GRANT ALL PERMISSIONS ON KEYSPACE messenger_keyspace TO messenger_app;
ALTER USER cassandra WITH PASSWORD 'DISABLED_SECURE_RAND_HASH_123_qwert';
