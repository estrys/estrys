CREATE TABLE users (
    username VARCHAR(15) PRIMARY KEY,
    id NUMERIC NOT NULL,
    private_key bytea NOT NULL ,
    created_at TIMESTAMP NOT NULL
)
