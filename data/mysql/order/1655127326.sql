CREATE TABLE `order_item`
(
    id       BINARY(16) PRIMARY KEY,
    order_id BINARY(16),
    price    BIGINT,
    quantity INT,
    FOREIGN KEY (order_id) REFERENCES `order` (id)
) ENGINE = InnoDB
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci