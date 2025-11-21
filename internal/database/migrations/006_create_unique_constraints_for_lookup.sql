-- Create unique constraints for customer_profile_lookup table
-- These constraints will enforce uniqueness at the database level
-- This allows us to INSERT directly and catch duplicates via database errors

-- For Individual customers with BVN: (firstName, surname, dateOfBirth, bvn)
-- This ensures uniqueness when BVN is provided
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_individual_bvn_unique
ON customer_profile_lookup("firstName", "surname", "dateOfBirth", "bvn")
WHERE "bvn" IS NOT NULL AND "firstName" IS NOT NULL AND "surname" IS NOT NULL AND "dateOfBirth" IS NOT NULL;

-- For Individual customers with NIN (but no BVN): (firstName, surname, dateOfBirth, nin)
-- This ensures uniqueness when NIN is provided but BVN is not
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_individual_nin_unique
ON customer_profile_lookup("firstName", "surname", "dateOfBirth", "nin")
WHERE "nin" IS NOT NULL AND "bvn" IS NULL AND "firstName" IS NOT NULL AND "surname" IS NOT NULL AND "dateOfBirth" IS NOT NULL;

-- For Individual customers without BVN/NIN: (firstName, surname, dateOfBirth)
-- This ensures uniqueness when neither BVN nor NIN is provided
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_individual_name_dob_unique
ON customer_profile_lookup("firstName", "surname", "dateOfBirth")
WHERE "bvn" IS NULL AND "nin" IS NULL AND "firstName" IS NOT NULL AND "surname" IS NOT NULL AND "dateOfBirth" IS NOT NULL;

-- For SME customers: (companyNameBusiness)
-- Company name must be unique (case-insensitive)
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_sme_company_unique
ON customer_profile_lookup(LOWER("companyNameBusiness"))
WHERE "companyNameBusiness" IS NOT NULL;

-- Note: These unique indexes will cause INSERT to fail with a unique constraint violation
-- if a duplicate exists, eliminating the need for SELECT before INSERT

