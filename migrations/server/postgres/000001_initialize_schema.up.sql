CREATE TABLE entries
(
    id         uuid UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    public_key bytea       NOT NULL,
    payload    bytea       NOT NULL
);