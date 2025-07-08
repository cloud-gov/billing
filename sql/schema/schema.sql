CREATE TABLE transactions (
  id SERIAL NOT NULL PRIMARY KEY,
  transaction_date DATE,
  resource_id INT NOT NULL,
  cf_org_id UUID NOT NULL,
  description TEXT,
  direction INT,
  amount INT NOT NULL,
  transaction_type_id INT NOT NULL,
  CONSTRAINT fk_resource_id Foreign Key (resource_id) REFERENCES resource(id),
  CONSTRAINT fk_cf_org_id Foreign Key (cf_org_id) REFERENCES cf_org(id),
  CONSTRAINT fk_transaction_type_id Foreign Key (transaction_type_id) REFERENCES transaction_type(id)
);

CREATE TABLE transaction_type (
  id SERIAL NOT NULL,
  name TEXT NOT NULL
);
