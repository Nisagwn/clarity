-- pg_trgm: tanınmayan alerjen terimleri için "şunu mu demek istediniz?"
-- önerileri benzerlik üzerinden hesaplanır.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_allergen_alias_trgm
    ON allergen_alias USING GIN (LOWER(alias) gin_trgm_ops);
