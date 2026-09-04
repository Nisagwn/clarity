-- Beauty Ingredient Explorer — Veritabanı şeması
--
-- skin_type gibi sabit değerler API sözleşmesinin parçası olduğu için
-- İngilizce kalır; arayüz bunları Türkçe etiketlerle gösterir.

CREATE TABLE IF NOT EXISTS ingredients (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    inci_name VARCHAR(255),
    description TEXT,
    concern_level INT DEFAULT 0, -- 1-10 endişe seviyesi (000005 ile türetilmiş hale geldi)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ingredient_allergens (
    id SERIAL PRIMARY KEY,
    ingredient_id INT NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    allergen_name VARCHAR(255) NOT NULL,
    severity INT, -- 1-10 şiddet seviyesi
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ingredient_id, allergen_name)
);

CREATE TABLE IF NOT EXISTS ingredient_benefits (
    id SERIAL PRIMARY KEY,
    ingredient_id INT NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    benefit VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ingredient_skin_types (
    id SERIAL PRIMARY KEY,
    ingredient_id INT NOT NULL REFERENCES ingredients(id) ON DELETE CASCADE,
    skin_type VARCHAR(50) NOT NULL, -- oily, dry, combination, sensitive, normal (API değerleri)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ingredient_id, skin_type)
);

CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    brand VARCHAR(255) NOT NULL,
    gtin VARCHAR(50), -- ürün takibi için barkod
    price DECIMAL(10, 2),
    currency VARCHAR(3) DEFAULT 'USD',
    image_url TEXT,
    category VARCHAR(100), -- fondöten, ruj, temizleyici vb.
    description TEXT,
    source_url TEXT, -- verinin alındığı kaynak (Sephora vb.)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(gtin, brand)
);

CREATE TABLE IF NOT EXISTS product_ingredients (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    ingredient_id INT NOT NULL REFERENCES ingredients(id),
    percentage DECIMAL(5, 2), -- isteğe bağlı: biliniyorsa
    order_index INT, -- INCI listesindeki sıra
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS user_profiles (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE,
    skin_type VARCHAR(50), -- oily, dry, combination, sensitive, normal (API değerleri)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_allergens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    allergen_name VARCHAR(255) NOT NULL,
    severity INT, -- 1-10 şiddet seviyesi
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, allergen_name)
);

CREATE TABLE IF NOT EXISTS user_favorites (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, product_id)
);

CREATE TABLE IF NOT EXISTS product_reviews (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    user_id INT,
    rating INT CHECK (rating >= 1 AND rating <= 5),
    review_text TEXT,
    flagged_allergens TEXT[], -- hangi içerikler sorun çıkardı?
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS price_history (
    id SERIAL PRIMARY KEY,
    product_id INT NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    price DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3),
    platform VARCHAR(100), -- Sephora, Ulta vb.
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Performans için indeksler
CREATE INDEX IF NOT EXISTS idx_ingredients_name ON ingredients(name);
CREATE INDEX IF NOT EXISTS idx_products_brand ON products(brand);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
CREATE INDEX IF NOT EXISTS idx_product_ingredients_product ON product_ingredients(product_id);
CREATE INDEX IF NOT EXISTS idx_product_ingredients_ingredient ON product_ingredients(ingredient_id);
CREATE INDEX IF NOT EXISTS idx_user_allergens_user ON user_allergens(user_id);
CREATE INDEX IF NOT EXISTS idx_price_history_product ON price_history(product_id);
CREATE INDEX IF NOT EXISTS idx_price_history_created ON price_history(created_at DESC);
