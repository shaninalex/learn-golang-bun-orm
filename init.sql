-- create custom schema to demonstrate connection to different schema then "public"
CREATE SCHEMA "warehouse";

-- make uuids available (for all schemas)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" schema pg_catalog;
