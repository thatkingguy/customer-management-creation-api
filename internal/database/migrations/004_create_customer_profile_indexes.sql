-- Create indexes for customer_profile table to optimize duplicate lookup queries
-- These indexes are critical for performance

-- Index for Individual customer lookup (firstName, surname, dateOfBirth)
CREATE INDEX IF NOT EXISTS idx_customer_profile_individual_lookup 
ON customer_profile("firstName", "surname", "dateOfBirth");

-- Index for BVN lookup (most unique identifier)
CREATE INDEX IF NOT EXISTS idx_customer_profile_bvn 
ON customer_profile("bvn") 
WHERE "bvn" IS NOT NULL;

-- Index for NIN lookup
CREATE INDEX IF NOT EXISTS idx_customer_profile_nin 
ON customer_profile("nin") 
WHERE "nin" IS NOT NULL;

-- Index for phone number lookup
CREATE INDEX IF NOT EXISTS idx_customer_profile_mobile_number 
ON customer_profile("mobileNumber") 
WHERE "mobileNumber" IS NOT NULL;

-- Index for SME customer lookup (companyNameBusiness)
CREATE INDEX IF NOT EXISTS idx_customer_profile_company_name 
ON customer_profile(LOWER("companyNameBusiness")) 
WHERE "companyNameBusiness" IS NOT NULL;

-- Index for tax identification number
CREATE INDEX IF NOT EXISTS idx_customer_profile_tax_id 
ON customer_profile(LOWER("taxIdentificationNumber")) 
WHERE "taxIdentificationNumber" IS NOT NULL;

-- Composite index for BVN + name lookup (optimized for FindDuplicateIndividual with BVN)
CREATE INDEX IF NOT EXISTS idx_customer_profile_bvn_name 
ON customer_profile("bvn", "firstName", "surname", "dateOfBirth") 
WHERE "bvn" IS NOT NULL;

-- Composite index for NIN + name lookup (optimized for FindDuplicateIndividual with NIN)
CREATE INDEX IF NOT EXISTS idx_customer_profile_nin_name 
ON customer_profile("nin", "firstName", "surname", "dateOfBirth") 
WHERE "nin" IS NOT NULL;

