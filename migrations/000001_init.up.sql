CREATE TABLE users (
    username VARCHAR(15) PRIMARY KEY,
    private_key bytea NOT NULL ,
    created_at TIMESTAMP NOT NULL
)