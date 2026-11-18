-- Example table backing the product CRUD module.
-- Naming: singular snake_case table, snake_case columns, UUID primary key,
-- audit columns created_at/created_by_name/updated_at/updated_by_name on every table.
CREATE TABLE IF NOT EXISTS product (
    id          UUID           PRIMARY KEY,
    code        VARCHAR(50)    NOT NULL,
    name        VARCHAR(255)   NOT NULL,
    description TEXT           NULL,
    price       NUMERIC(18, 2) NOT NULL DEFAULT 0 CHECK (price >= 0),
    status      VARCHAR(20)    NOT NULL CHECK (status IN ('ACTIVE', 'INACTIVE')),
    created_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    created_by_name  VARCHAR(255)   NOT NULL DEFAULT '',
    updated_at  TIMESTAMPTZ    NOT NULL DEFAULT now(),
    updated_by_name  VARCHAR(255)   NOT NULL DEFAULT ''
);

-- The usecase checks uniqueness first for a friendly 409; this constraint is
-- the real guarantee under concurrent requests.
CREATE UNIQUE INDEX IF NOT EXISTS ux_product_code ON product (code);
CREATE INDEX IF NOT EXISTS ix_product_status ON product (status);
CREATE INDEX IF NOT EXISTS ix_product_created_at ON product (created_at DESC);
