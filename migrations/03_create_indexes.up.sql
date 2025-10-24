CREATE INDEX idx_clicks_alias_clicked_at ON clicks(alias, clicked_at);
CREATE INDEX idx_clicks_alias_client ON clicks(alias, client_name);
CREATE INDEX idx_clicks_alias_device ON clicks(alias, device_type);