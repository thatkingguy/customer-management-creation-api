-- Create indexes for request and interim_approval_config tables to optimize queries
-- These indexes are critical for performance

-- Note: requestId is typically the primary key, so it already has an index
-- But we add a partial index for non-deleted records to optimize the WHERE clause
CREATE INDEX IF NOT EXISTS idx_request_request_id_not_deleted 
ON request("requestId") 
WHERE "isDeleted" = false;

-- Composite index for request status filtering (if needed for other queries)
CREATE INDEX IF NOT EXISTS idx_request_status_deleted 
ON request("status", "isDeleted") 
WHERE "isDeleted" = false;

-- CRITICAL: Composite index for interim_approval_config lookup
-- This index covers the WHERE clause (customerType, gracePeriod, status) and ORDER BY (createdAt DESC)
-- PostgreSQL can use this index to answer the query without scanning the table
CREATE INDEX IF NOT EXISTS idx_interim_approval_config_lookup 
ON interim_approval_config("customerType", "gracePeriod", "status", "createdAt" DESC);

-- Alternative partial index for active interim approval configs only (more efficient if table is large)
-- This index only includes rows where gracePeriod=true AND status='Approved'
CREATE INDEX IF NOT EXISTS idx_interim_approval_config_active 
ON interim_approval_config("customerType", "createdAt" DESC) 
WHERE "gracePeriod" = true AND "status" = 'Approved';

