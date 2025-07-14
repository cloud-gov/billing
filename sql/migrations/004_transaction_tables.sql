CREATE TABLE transaction_type (
  id SERIAL NOT NULL PRIMARY KEY,
  name TEXT NOT NULL
);

CREATE TABLE transaction (
  id SERIAL NOT NULL PRIMARY KEY,
  transaction_date DATE,
  cf_org_id UUID NOT NULL,
  description TEXT,
  direction INT,
  amount INT NOT NULL,
  transaction_type_id INT NOT NULL,
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_transaction_type_id Foreign Key (transaction_type_id) REFERENCES transaction_type(id)
);

---- create above / drop below ----

DROP TABLE transaction;
DROP TABLE transaction_type;
