CREATE TABLE `message`
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    type       VARCHAR(255) NOT NULL,
    topic      TEXT         NOT NULL,
    `key`      TEXT         NOT NULL,
    body       BLOB         NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci