CREATE TABLE  IF NOT EXISTS url_mapping (
    uuid UUID PRIMARY KEY,
    short_url VARCHAR(255) NOT NULL,
    original_url TEXT NOT NULL
);