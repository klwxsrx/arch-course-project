CREATE TABLE `product`
(
    id          BINARY(16) PRIMARY KEY,
    title       TEXT,
    description MEDIUMTEXT,
    price       BIGINT,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci