CREATE TABLE actors (
    id uuid PRIMARY KEY,
    url VARCHAR(2048) NOT NULL UNIQUE,
    public_key bytea NOT NULL
)