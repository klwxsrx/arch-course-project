CREATE TABLE `user`
(
    id               BINARY(16) PRIMARY KEY,
    login            VARCHAR(255),
    encoded_password CHAR(60) BINARY,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci