-- Stored grid assets. The hierarchy is a self-reference on the natural key
-- asset_id (research R7). Only PK / unique / NOT NULL / FK constraints: asset
-- types, statuses, the permitted-parent rules and rating_kva >= 0 are enforced
-- by the backend (constitution II, Template Exception 8). Blank CSV values are
-- stored as NULL. Audit names are set by the usecase ("system").
CREATE TABLE IF NOT EXISTS asset (
    id                 UUID         PRIMARY KEY,
    asset_id           TEXT         NOT NULL,
    parent_asset_id    TEXT         NULL,
    asset_type         VARCHAR(32)  NOT NULL,
    asset_name         TEXT         NOT NULL,
    operational_status VARCHAR(32)  NULL,
    voltage_kv         NUMERIC      NULL,
    rating_kva         NUMERIC      NULL,
    manufacturer       TEXT         NULL,
    model              TEXT         NULL,
    serial_number      TEXT         NULL,
    commissioned_date  DATE         NULL,
    import_id          UUID         NOT NULL,
    created_at         TIMESTAMPTZ  NOT NULL,
    created_by_name    VARCHAR(255) NOT NULL,
    updated_at         TIMESTAMPTZ  NOT NULL,
    updated_by_name    VARCHAR(255) NOT NULL,
    CONSTRAINT fk_asset_import FOREIGN KEY (import_id) REFERENCES asset_import (id)
);

-- Unique natural key; also the target of the parent foreign key below.
CREATE UNIQUE INDEX IF NOT EXISTS ux_asset_asset_id ON asset (asset_id);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_asset_parent') THEN
        ALTER TABLE asset
            ADD CONSTRAINT fk_asset_parent FOREIGN KEY (parent_asset_id)
            REFERENCES asset (asset_id) ON DELETE RESTRICT;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS ix_asset_parent_asset_id ON asset (parent_asset_id);
CREATE INDEX IF NOT EXISTS ix_asset_serial_number ON asset (serial_number);
CREATE INDEX IF NOT EXISTS ix_asset_import_id ON asset (import_id);
