CREATE TYPE account_order_event_status AS ENUM ('PAID', 'CANCELED');

CREATE TABLE IF NOT EXISTS accounts (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT UNIQUE NOT NULL,
  amount_cents BIGINT NOT NULL
);

CREATE INDEX account_user_id_idx ON accounts (user_id);

INSERT INTO accounts (user_id, amount_cents)
  VALUES (1, 100000),
    (2, 200000),
    (3, 300000),
    (4, 400000),
    (5, 500000);

CREATE TABLE IF NOT EXISTS account_events (
  id BIGSERIAL PRIMARY KEY,
  account_id BIGINT REFERENCES accounts ON DELETE SET NULL,
  order_id BIGINT,
  order_event_id BIGINT UNIQUE NOT NULL,
  status ACCOUNT_ORDER_EVENT_STATUS NOT NULL
);

ALTER TABLE account_events REPLICA IDENTITY FULL;

CREATE PUBLICATION account_events_publication FOR TABLE account_events;

SELECT pg_create_logical_replication_slot('account_events_replication', 'pgoutput');
