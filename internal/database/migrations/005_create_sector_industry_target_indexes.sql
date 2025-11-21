-- Create indexes for sector, industry, and target tables to optimize lookup queries
-- These indexes are critical for performance

-- Index for sector lookup by sectorId (primary key lookup)
CREATE INDEX IF NOT EXISTS idx_sector_sector_id 
ON sector("sectorId");

-- Index for sector lookup by description (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_sector_description_lower 
ON sector(LOWER(description));

-- Index for industry lookup by industryId (primary key lookup)
CREATE INDEX IF NOT EXISTS idx_industry_industry_id 
ON industry("industryId");

-- Index for industry lookup by description (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_industry_description_lower 
ON industry(LOWER(description));

-- Index for industry lookup by sectorId (for validation)
CREATE INDEX IF NOT EXISTS idx_industry_sector_id 
ON industry("sectorId");

-- Index for target lookup by targetId (primary key lookup)
CREATE INDEX IF NOT EXISTS idx_target_target_id 
ON target("targetId");

-- Index for target lookup by description (case-insensitive)
CREATE INDEX IF NOT EXISTS idx_target_description_lower 
ON target(LOWER(description));

