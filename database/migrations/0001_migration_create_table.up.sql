BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE coupons (
    id UUID                     PRIMARY KEY DEFAULT gen_random_uuid(),
    coupon_code                 TEXT UNIQUE NOT NULL,
    expiry_date                 TIMESTAMP NOT NULL,
    usage_type                  TEXT CHECK (usage_type IN ('one_time', 'multi_use', 'time_based')) NOT NULL,
    min_order_value             NUMERIC DEFAULT 0,
    valid_from                  TIMESTAMP,
    valid_to                    TIMESTAMP,
    terms_and_conditions        TEXT,
    discount_type               TEXT CHECK (discount_type IN ('percentage', 'fixed')) NOT NULL,
    discount_value              NUMERIC NOT NULL,
    max_usage_per_user          INT DEFAULT 1,
    target                      TEXT CHECK (target IN ('inventory', 'charges')) NOT NULL,
    created_at                  TIMESTAMP DEFAULT NOW(),
    updated_at                  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE coupon_applicable_medicines (
    id                SERIAL      PRIMARY KEY,
    coupon_id         UUID REFERENCES coupons(id) ON DELETE CASCADE,
    medicine_id       TEXT NOT NULL
);

CREATE TABLE coupon_applicable_categories (
    id                   SERIAL PRIMARY KEY,
    coupon_id            UUID REFERENCES coupons(id) ON DELETE CASCADE,
    category             TEXT NOT NULL
);

CREATE TABLE coupon_usages (
    id                   SERIAL PRIMARY KEY,
    coupon_id            UUID REFERENCES coupons(id) ON DELETE CASCADE,
    user_id              TEXT NOT NULL,
    used_at              TIMESTAMP DEFAULT NOW(),
    UNIQUE (coupon_id)
);

COMMIT;
