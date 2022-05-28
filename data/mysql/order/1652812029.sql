CREATE TABLE `order`
(
    id           BINARY(16) PRIMARY KEY,
    user_id      BINARY(16),
    address_id   BINARY(16),
    status       TINYINT,
    total_amount BIGINT,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci