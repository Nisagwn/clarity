-- Gerçek katalog: kaynak, lisans ve veri kalitesi
--
-- Neden: katalogdaki 14 ürün kurgusaldı. Gerçek veri Open Beauty Facts'ten
-- geliyor ve ODbL lisanslı; ODbL atıf ve aynı lisansla paylaşma gerektiriyor.
-- Atıf yapabilmek için her ürünün nereden geldiğini KAYDETMEK zorundayız:
-- lisans yükümlülüğü, veriyi gösterdiğimiz her yerde geçerli.
--
-- İkinci neden: crowdsource veri eksik olabilir. İçerik listesi olmayan ya da
-- üç içerikten az listelenen bir ürünle Jaccard benzerliği hesaplamak
-- yanıltıcı sonuç üretir — iki ürünün "aynı" görünmesi, ikisinin de listesinin
-- boş olmasından kaynaklanabilir. Bu yüzden veri kalitesi bir kolon.

ALTER TABLE products
    ADD COLUMN source       VARCHAR(50),   -- 'openbeautyfacts', 'seed', ...
    ADD COLUMN source_id    VARCHAR(100),  -- kaynaktaki kimlik (OBF'de barkod)
    ADD COLUMN license      VARCHAR(50),   -- 'ODbL-1.0'
    ADD COLUMN verified_at  TIMESTAMP,     -- kaynaktan en son ne zaman çekildi
    ADD COLUMN data_quality VARCHAR(20) NOT NULL DEFAULT 'ok';

ALTER TABLE products
    ADD CONSTRAINT products_data_quality_check
    CHECK (data_quality IN ('ok', 'incomplete'));

-- İçe aktarım idempotent olsun: aynı kaynak kimliği ikinci kez ürün açmasın.
CREATE UNIQUE INDEX idx_products_source
    ON products (source, source_id)
    WHERE source IS NOT NULL AND source_id IS NOT NULL;

-- Muadil sorgusu eksik veriyi eleyecek; kolon süzgeçte kullanılıyor.
CREATE INDEX idx_products_data_quality ON products (data_quality);

-- İçerikler de kaynağını taşır: içe aktarımda katalogda karşılığı olmayan
-- INCI adları aday olarak eklenir. Bunlar küratörlükten geçmemiştir ve
-- mevzuat eşleşmesi olmadığı sürece puansız kalır (score_version NULL).
ALTER TABLE ingredients
    ADD COLUMN source VARCHAR(50);

-- Mevcut veri kurgusal örnek veriydi; olduğu gibi işaretlensin ki gerçek
-- kayıtlarla karışmasın.
UPDATE products SET source = 'seed' WHERE source IS NULL;
UPDATE ingredients SET source = 'seed' WHERE source IS NULL;
