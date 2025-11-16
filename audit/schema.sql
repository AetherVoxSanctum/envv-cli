CREATE TABLE audit_event (
  id SERIAL PRIMARY KEY,
  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  action TEXT,
  username TEXT,
  file TEXT
);

-- Create the sops role with a secure password
-- Replace 'YOUR_SECURE_PASSWORD' with an actual secure password
-- Or use: CREATE ROLE sops WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN;
-- And then set the password separately: ALTER ROLE sops WITH PASSWORD 'your_secure_password';
CREATE ROLE sops WITH NOSUPERUSER INHERIT NOCREATEROLE NOCREATEDB LOGIN PASSWORD 'YOUR_SECURE_PASSWORD';

GRANT INSERT ON audit_event TO sops;
GRANT USAGE ON audit_event_id_seq TO sops;
