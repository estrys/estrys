CREATE TABLE followers (
    "user" VARCHAR(15) NOT NULL REFERENCES users(username) ON DELETE CASCADE,
    actor uuid NOT NULL REFERENCES actors(id) ON DELETE CASCADE,
    PRIMARY KEY ("user", actor)
)