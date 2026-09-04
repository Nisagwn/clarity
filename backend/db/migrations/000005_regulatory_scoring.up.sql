-- Mevzuat verisi ve puanlama rubriği
--
-- Neden: concern_level değerleri elle atanmıştı. EWG kendi puanlarının
-- izinsiz kullanılmasını yasaklıyor ve herkese açık bir API'si yok; yani
-- "EWG puanı göstermek" lisans olmadan mümkün değil.
--
-- Çözüm: puanı AB Tüzüğü 1223/2009 Eklerinden TÜRETMEK ve her puanın
-- kaynağını göstermek. Bu hem hukuken güvenli hem kullanıcı için daha
-- dürüst: puanın nereden geldiği görünür oluyor.

CREATE TABLE ingredient_regulatory (
    ingredient_id     INT PRIMARY KEY REFERENCES ingredients(id) ON DELETE CASCADE,
    cas_number        VARCHAR(40),
    ec_number         VARCHAR(40),
    -- 'II' yasaklı, 'III' kısıtlı, 'IV' renklendirici, 'V' koruyucu, 'VI' UV filtresi
    annex             VARCHAR(10),
    annex_entry       VARCHAR(40),
    restriction       TEXT,
    max_concentration NUMERIC(6,3),
    -- Ek III'te yer alan ve etikette beyanı zorunlu koku alerjeni mi
    declarable_allergen BOOLEAN NOT NULL DEFAULT FALSE,
    sccs_opinion      TEXT,
    sccs_adverse      BOOLEAN NOT NULL DEFAULT FALSE,
    source_url        TEXT NOT NULL,
    fetched_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ingredient_regulatory_annex ON ingredient_regulatory (annex);

-- Puanlama kuralları kodda değil veritabanında ve versiyonlu tutulur:
-- rubrik değişince hangi puanın hangi kurala göre verildiği izlenebilir.
CREATE TABLE scoring_rule (
    id        SERIAL PRIMARY KEY,
    version   INT NOT NULL,
    rule_key  VARCHAR(40) NOT NULL,
    score     INT NOT NULL CHECK (score BETWEEN 1 AND 10),
    rationale TEXT NOT NULL,
    UNIQUE(version, rule_key)
);

ALTER TABLE ingredients
    ADD COLUMN score_version INT,
    ADD COLUMN score_sources TEXT[];

-- concern_level artık türetilmiş bir değer. Mevzuat verisi olmayan
-- içeriklerde NULL kalır: uydurma bir sayı göstermektense "henüz
-- puanlanmadı" demek dürüst olan.
ALTER TABLE ingredients ALTER COLUMN concern_level DROP DEFAULT;

-- ===== Rubrik v1 =====

INSERT INTO scoring_rule (version, rule_key, score, rationale) VALUES
    (1, 'annex_ii_banned',        10, 'AB''de kozmetik ürünlerde kullanımı yasak (Ek II)'),
    (1, 'declarable_allergen',     7, 'Etikette beyanı zorunlu tutulan temas alerjeni (Ek III)'),
    (1, 'annex_iii_restricted',    5, 'Belirli bir oranın üstünde güvenli kabul edilmiyor (Ek III)'),
    (1, 'annex_v_preservative',    4, 'Koruyucu olarak koşullu izinli (Ek V)'),
    (1, 'annex_vi_uv_filter',      4, 'UV filtresi olarak koşullu izinli (Ek VI)'),
    (1, 'annex_iv_colorant',       3, 'Renklendirici olarak izinli, kullanım alanı kısıtlı olabilir (Ek IV)'),
    (1, 'unrestricted',            2, 'Eklerde kısıtlama kaydı yok'),
    (1, 'sccs_adverse_modifier',   2, 'Bilimsel Tüketici Güvenliği Komitesi olumsuz görüş bildirdi (puana eklenir)');

-- ===== Mevcut katalog için mevzuat verisi =====
--
-- Yalnızca TANIMI GEREĞİ kesin olan kayıtlar girilmiştir:
--   * Ek III bildirimli koku alerjenleri — bildirimli olmaları zaten
--     Ek III'te yer almalarından kaynaklanır
--   * Ek V koruyucular, Ek VI UV filtreleri, Ek IV renklendiriciler —
--     ürün içindeki işlevleri bu Ek'leri belirler
--
-- Geri kalan içerikler BİLİNÇLİ olarak boş bırakılmıştır. Tam CosIng
-- dışa aktarımı içe alınana kadar (backend/scripts/import-cosing)
-- puanları NULL kalır ve arayüzde "henüz puanlanmadı" görünür.
-- Tahmin edilmiş bir Ek numarası, tahmin edilmiş bir puandan daha kötüdür:
-- mevzuata atıf yapıyormuş gibi görünüp yanlış olur.

INSERT INTO ingredient_regulatory
    (ingredient_id, annex, annex_entry, declarable_allergen, restriction, source_url)
SELECT i.id, v.annex, v.entry, v.declarable, v.restriction, v.url
FROM (VALUES
    -- Ek III — bildirimli koku alerjenleri
    ('Linalool',          'III', '84', TRUE,
     'Etikette beyanı zorunlu; yıkanan üründe %0,001, yıkanmayan üründe %0,01 üstünde bildirilir',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Limonen',           'III', '89', TRUE,
     'Etikette beyanı zorunlu; peroksit değeri sınırlı',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Benzil Salisilat',  'III', '76', TRUE,
     'Etikette beyanı zorunlu',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),

    -- Ek V — koruyucular
    ('Fenoksietanol',     'V',   '29', FALSE,
     'Koruyucu olarak en fazla %1',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Metilparaben',      'V',   '12', FALSE,
     'Koruyucu olarak kısıtlı; tek başına ve toplam paraben sınırı var',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('DMDM Hidantoin',    'V',   '33', FALSE,
     'Formaldehit salıcı koruyucu; salınan formaldehit için etiket uyarısı gerekir',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),

    -- Ek VI — UV filtreleri
    ('Titanyum Dioksit',  'VI',  '27', FALSE,
     'UV filtresi olarak en fazla %25; solunabilir toz biçimi ayrı değerlendirilir',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Çinko Oksit',       'VI',  '30', FALSE,
     'UV filtresi olarak en fazla %25',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),

    -- Ek IV — renklendiriciler
    ('Demir Oksitler',    'IV',  'CI 77491/77492/77499', FALSE,
     'Renklendirici olarak izinli',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Mika',              'IV',  'CI 77019', FALSE,
     'Renklendirici olarak izinli',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('Karmin',            'IV',  'CI 75470', FALSE,
     'Renklendirici olarak izinli',
     'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223')
) AS v(ing_name, annex, entry, declarable, restriction, url)
JOIN ingredients i ON i.name = v.ing_name;
