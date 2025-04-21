CREATE DATABASE go_auth_db;

\c go_auth_db;

DROP TABLE IF EXISTS accounts;

CREATE TABLE accounts (
    account_id BIGSERIAL PRIMARY KEY,
    account_name VARCHAR NOT NULL DEFAULT '',
    account_email VARCHAR NOT NULL DEFAULT '',
    account_phone_number VARCHAR NOT NULL DEFAULT '',
    account_password VARCHAR NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT
);

CREATE TABLE refresh_tokens (
    refresh_token_id BIGSERIAL PRIMARY KEY,
    refresh_token VARCHAR NOT NULL DEFAULT '',
    account_id BIGINT NOT NULL,
    device_id BIGINT NOT NULL,
    expired_at BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL
);

CREATE TABLE account_devices (
    device_id BIGSERIAL PRIMARY KEY,
    account_id BIGINT NOT NULL,
    device_hash VARCHAR NOT NULL,
    user_agent VARCHAR NOT NULL,
    device_info VARCHAR NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT NOT NULL,
    deleted_at BIGINT
);