-- 000001_init geri alma
--
-- Bağımlılık sırası önemli: yabancı anahtarla bağlı tablolar önce düşer.

DROP INDEX IF EXISTS idx_price_history_created;
DROP INDEX IF EXISTS idx_price_history_product;
DROP INDEX IF EXISTS idx_user_allergens_user;
DROP INDEX IF EXISTS idx_product_ingredients_ingredient;
DROP INDEX IF EXISTS idx_product_ingredients_product;
DROP INDEX IF EXISTS idx_products_category;
DROP INDEX IF EXISTS idx_products_brand;
DROP INDEX IF EXISTS idx_ingredients_name;

DROP TABLE IF EXISTS price_history;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS user_favorites;
DROP TABLE IF EXISTS user_allergens;
DROP TABLE IF EXISTS user_profiles;
DROP TABLE IF EXISTS product_ingredients;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS ingredient_skin_types;
DROP TABLE IF EXISTS ingredient_benefits;
DROP TABLE IF EXISTS ingredient_allergens;
DROP TABLE IF EXISTS ingredients;
