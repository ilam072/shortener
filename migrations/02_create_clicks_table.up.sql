CREATE TABLE IF NOT EXISTS clicks (
                                      id UUID PRIMARY KEY,
                                      alias TEXT REFERENCES links(alias),
                                      clicked_at TIMESTAMP DEFAULT now(),
                                      user_agent TEXT,
                                      client_name TEXT,
                                      device_type TEXT,
                                      ip inet
);