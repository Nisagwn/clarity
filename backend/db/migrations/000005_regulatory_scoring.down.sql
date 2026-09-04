ALTER TABLE ingredients
    DROP COLUMN IF EXISTS score_sources,
    DROP COLUMN IF EXISTS score_version;
ALTER TABLE ingredients ALTER COLUMN concern_level SET DEFAULT 0;
DROP TABLE IF EXISTS scoring_rule;
DROP INDEX IF EXISTS idx_ingredient_regulatory_annex;
DROP TABLE IF EXISTS ingredient_regulatory;
