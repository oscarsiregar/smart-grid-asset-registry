-- One row per completed confirm attempt of a CSV import (spec FR-029).
-- Only PK / NOT NULL constraints: every business rule (mode and outcome
-- values, count invariant) is enforced by the backend (constitution II,
-- Template Exception 8). Audit names are set by the usecase ("system").
CREATE TABLE IF NOT EXISTS asset_import (
    id              UUID         PRIMARY KEY,
    file_name       TEXT         NOT NULL,
    mode            VARCHAR(16)  NOT NULL,
    outcome         VARCHAR(32)  NOT NULL,
    total_rows      INTEGER      NOT NULL,
    imported_rows   INTEGER      NOT NULL,
    unchanged_rows  INTEGER      NOT NULL,
    rejected_rows   INTEGER      NOT NULL,
    warning_rows    INTEGER      NOT NULL,
    created_at      TIMESTAMPTZ  NOT NULL,
    created_by_name VARCHAR(255) NOT NULL,
    updated_at      TIMESTAMPTZ  NOT NULL,
    updated_by_name VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS ix_asset_import_created_at ON asset_import (created_at DESC);
