-- Runs once, when the db volume is first initialised (docker-entrypoint-initdb.d).
-- Integration tests use their own database so they can truncate tables
-- without touching the data in asset_registry (research R13).
CREATE DATABASE asset_registry_test;
