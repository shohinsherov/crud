    CREATE TABLE users (
        id BIGSERIAL PRIMARY KEY,
        name    TEXT   NOT NULL,
        phone   TEXT   NOT NULL UNIQUE,
        password TEXT  NOT NULL,
        roles      TEXT[] NOT NULL DEFAULT '{}',
        active BOOLEAN NOT NULL DEFAULT TRUE,
        created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP  
    )