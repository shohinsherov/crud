CREATE TABLE customers (
id BIGSERIAL PRIMARY KEY,
name	TEXT NOT NULL,
phone 	text 	NOT NULL UNIQUE,
password TEXT 	NOT NULL,
active 	BOOLEAN NOT NULL DEFAULT TRUE,
created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP 
);
