CREATE TABLE IF NOT EXISTS links (
                        id UUID PRIMARY KEY,
                        url TEXT NOT NULL,
                        alias TEXT UNIQUE NOT NULL,
                        created_at TIMESTAMP DEFAULT now()
);