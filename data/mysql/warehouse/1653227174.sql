CREATE TABLE `stock_balance`
(
    id         BINARY(16) PRIMARY KEY,
    item_id    BINARY(16),
    type       TINYINT,
    quantity   INT,
    order_id   BINARY(16) DEFAULT NULL,
    created_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP  DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP  DEFAULT NULL,
    INDEX order_id_index (order_id),
    INDEX item_index (item_id, deleted_at)
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci