-- Second database for Go integration tests (keeps keystone DB free for manual gostone runs).
CREATE DATABASE IF NOT EXISTS gostone_integration
  CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
GRANT ALL PRIVILEGES ON gostone_integration.* TO 'keystone'@'%';
FLUSH PRIVILEGES;
