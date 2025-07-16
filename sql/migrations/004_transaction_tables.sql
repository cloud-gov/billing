CREATE TYPE transaction_type AS ENUM (
  'iaa_pop_start',
  'iaa_pop_end',
  'usage_post'
);

COMMENT ON TYPE transaction_type IS 'TransactionType explains why the transaction was made. Each means:
  - iaa_pop_start: The IAA Period of Performance started.
  - iaa_pop_end: The IAA Period of Performance ended.
  - usage_post: Customer usage of was posted, i.e. their account balance was updated to reflect their usage.
';

CREATE TABLE account_type (
  id INT PRIMARY KEY,
  name TEXT NOT NULL,
  normal INT -- 1 or -1
);

CREATE TABLE account (
  id SERIAL PRIMARY KEY,
  customer_id BIGINT,
  type INT,

  CONSTRAINT fk_customer_id Foreign Key (customer_id) REFERENCES customer(id),
  CONSTRAINT fk_type_id Foreign Key (type) REFERENCES account_type(id)
);

CREATE UNIQUE INDEX account_unique ON account(id, type);

CREATE TABLE transaction (
  id SERIAL PRIMARY KEY,
  occurred_at TIMESTAMP,
  description TEXT,
  type transaction_type NOT NULL
);

CREATE TABLE entry (
  transaction_id INT REFERENCES transaction(id),
  account_id INT,
  amount NUMERIC(20,4) NOT NULL,
  direction INT, -- 1 or -1.

  PRIMARY KEY (transaction_id, account_id),
  CONSTRAINT fk_account_id Foreign Key (account_id) REFERENCES account(id)
);

CREATE OR REPLACE FUNCTION assert_transaction_balances()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    PERFORM 1
    FROM entry e
    GROUP BY e.transaction_id
    HAVING SUM(e.amount) <> 0;

    IF FOUND THEN
        RAISE EXCEPTION
          'ledger error: at least one transaction is not balanced (sum(amount) <> 0)';
    END IF;

    RETURN NULL;
END;
$$;

CREATE CONSTRAINT TRIGGER transaction_balances_chk
AFTER INSERT OR UPDATE OR DELETE ON entry
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE PROCEDURE assert_transaction_balances();

INSERT INTO account_type (id, name, normal) VALUES
(200, 'liabilities', -1),
(201, 'credit_pool', -1),
(400, 'expenses', 1),
(401, 'credits_used', 1);

---- create above / drop below ----

DROP TABLE entry;
DROP TABLE transaction;
DROP TYPE transaction_type;
DROP TABLE account;
DROP TABLE account_type;
