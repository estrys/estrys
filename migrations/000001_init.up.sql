CREATE TABLE users (
    username VARCHAR(15) PRIMARY KEY,
    id VARCHAR(20) NOT NULL,
    private_key bytea NOT NULL ,
    created_at TIMESTAMP NOT NULL
)
