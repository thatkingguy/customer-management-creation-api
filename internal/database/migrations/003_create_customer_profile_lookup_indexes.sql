-- Create indexes and unique constraints for customer_profile_lookup table
-- Note: These should match the actual schema requirements

-- Unique constraint for Individual customers (firstName, surname, bvn, dateOfBirth)
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_individual_bvn 
ON customer_profile_lookup("firstName", "surname", "bvn", "dateOfBirth") 
WHERE "bvn" IS NOT NULL;

-- Unique constraint for Individual customers (firstName, surname, nin, dateOfBirth)
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_individual_nin 
ON customer_profile_lookup("firstName", "surname", "nin", "dateOfBirth") 
WHERE "nin" IS NOT NULL;

-- Unique constraint for SME customers (companyNameBusiness, certificateOfIncorporation, taxIdentificationNumber, dateOfRegistration)
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_sme_full 
ON customer_profile_lookup("companyNameBusiness", "certificateOfIncorporation", "taxIdentificationNumber", "dateOfRegistration") 
WHERE "companyNameBusiness" IS NOT NULL;

-- Partial index for SME company name only
CREATE UNIQUE INDEX IF NOT EXISTS idx_customer_lookup_sme_company 
ON customer_profile_lookup("companyNameBusiness") 
WHERE "companyNameBusiness" IS NOT NULL;



