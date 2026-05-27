-- Initialize tiaozao database
-- This runs on first Postgres startup via docker-entrypoint-initdb.d

-- Set timezone
SET timezone = 'Asia/Shanghai';

-- Create extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Note: Tables will be created by GORM AutoMigrate or migration scripts
-- See backend/migrations/ for DDL

CREATE SCHEMA IF NOT EXISTS public;
