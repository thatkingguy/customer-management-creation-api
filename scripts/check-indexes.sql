-- Script to check if indexes exist and are being used
-- Run this to verify indexes are created: psql -h <host> -U <user> -d <db> -f scripts/check-indexes.sql

-- Check request table indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'request'
ORDER BY indexname;

-- Check interim_approval_config table indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'interim_approval_config'
ORDER BY indexname;

-- Check if primary key exists on request table
SELECT
    tc.table_name, 
    kcu.column_name,
    tc.constraint_name
FROM information_schema.table_constraints AS tc 
JOIN information_schema.key_column_usage AS kcu
  ON tc.constraint_name = kcu.constraint_name
WHERE tc.constraint_type = 'PRIMARY KEY' 
  AND tc.table_name = 'request';

-- Explain plan for request query (replace with actual UUID)
-- EXPLAIN ANALYZE SELECT * FROM request WHERE "requestId" = '00000000-0000-0000-0000-000000000000' AND "isDeleted" = false;

-- Explain plan for interim approval config query
-- EXPLAIN ANALYZE SELECT * FROM interim_approval_config WHERE "customerType" = 'Individual' AND "gracePeriod" = true AND "status" = 'Approved' ORDER BY "createdAt" DESC LIMIT 1;

