CREATE TYPE order_status AS ENUM ('CREATED', 'PAID', 'CANCELED');

CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status ORDER_STATUS NOT NULL
);

CREATE TABLE IF NOT EXISTS order_create_events (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT REFERENCES orders ON DELETE RESTRICT,
  amount_cents BIGINT NOT NULL,
  user_id BIGINT NOT NULL
);

ALTER TABLE order_create_events REPLICA IDENTITY FULL;

CREATE PUBLICATION order_events_publication FOR TABLE order_create_events;

SELECT pg_create_logical_replication_slot('order_events_replication', 'pgoutput');
