ALTER TABLE ingredients DROP COLUMN IF EXISTS source;

DROP INDEX IF EXISTS idx_products_data_quality;
DROP INDEX IF EXISTS idx_products_source;

ALTER TABLE products DROP CONSTRAINT IF EXISTS products_data_quality_check;
ALTER TABLE products
    DROP COLUMN IF EXISTS data_quality,
    DROP COLUMN IF EXISTS verified_at,
    DROP COLUMN IF EXISTS license,
    DROP COLUMN IF EXISTS source_id,
    DROP COLUMN IF EXISTS source;
