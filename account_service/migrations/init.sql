CREATE TYPE order_event_status AS ENUM ('PAID', 'CANCELED');

CREATE TABLE IF NOT EXISTS accounts (
  id BIGINT PRIMARY KEY AUTO INCREMENT,
  user_id BIGSERIAL NOT NULL,
  amount_cents BIGINT NOT NULL
);

INSERT INTO accounts 
  VALUES (
    (1, 100000),
    (2, 200000),
    (3, 300000),
    (4, 400000),
    (5, 500000),
  );

CREATE TABLE IF NOT EXISTS account_events (
  id BIGINT PRIMARY KEY AUTO INCREMENT,
  account_id BIGINT NOT NULL REFERENCES accounts ON DELETE RESTRICT,
  order_id BIGINT,
  order_event_id BIGINT NOT NULL UNIQUE,
  status ORDER_EVENT_STATUS NOT NULL
);

ALTER TABLE account_events REPLICA IDENTITY FULL;

CREATE PUBLICATION account_events_publication FOR TABLE account_events;

SELECT pg_create_logical_replication_slot('account_events_publication', 'pgoutput');
