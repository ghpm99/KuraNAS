CREATE TABLE IF NOT EXISTS activity_diary (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    description VARCHAR(256),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP
);