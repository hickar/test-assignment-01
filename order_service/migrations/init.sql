CREATE TYPE order_status AS ENUM ('CREATED', 'PAID', 'CANCELLED');

CREATE TABLE IF NOT EXISTS orders (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL,
  amount_cents BIGINT NOT NULL,
  status ORDER_STATUS NOT NULL
);

CREATE TABLE IF NOT EXISTS order_events (
  id BIGSERIAL PRIMARY KEY,
  order_id BIGINT REFERENCES orders ON DELETE RESTRICT
);

ALTER TABLE order_events REPLICA IDENTITY FULL;

CREATE PUBLICATION order_events_publication FOR TABLE order_events;

SELECT pg_create_logical_replication_slot('postgres_debezium', 'pgoutput');
