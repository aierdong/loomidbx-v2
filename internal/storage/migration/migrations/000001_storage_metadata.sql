-- 000001_storage_metadata.sql
-- 仅建立迁移记录基础结构，不创建连接、Schema、Project 或执行历史业务表。
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    checksum TEXT NOT NULL,
    applied_at INTEGER NOT NULL
);
