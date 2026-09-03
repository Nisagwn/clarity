-- 000002 geri alma
DROP INDEX IF EXISTS idx_allergen_alias_canonical;
DROP INDEX IF EXISTS idx_allergen_alias_lower;
DROP TABLE IF EXISTS allergen_alias;
DROP TABLE IF EXISTS allergen_canonical;
