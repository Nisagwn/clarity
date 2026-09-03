-- Kanonik alerjen sözlüğü
--
-- Neden: alerjen eşleştirmesi iki yönlü alt dize karşılaştırmasıyla yapılıyordu
-- (LIKE '%' || terim || '%'). Bu, "alkol" arayan kullanıcıyı alerjeni
-- "yün alkolü" olarak kayıtlı Lanolin'e takıyordu. Yün alkolleri sterol
-- yapıdadır, etanolle ilgisi yoktur. İnsanların reaksiyondan kaçınmak için
-- güvendiği bir özellikte yanlış uyarı, kaçırılan uyarı kadar zararlıdır.
--
-- Çözüm: hem kullanıcının girdisi hem içeriğin alerjen adı aynı sözlükten
-- TAM EŞLEŞMEYLE çözülür. Alt dize karşılaştırması tamamen kalkar.
--
-- Sözlük uydurma değil: koku alerjenleri AB Tüzüğü 1223/2009 Ek III'ten,
-- etikette beyanı zorunlu tutulan maddeler. Bunlar mevzuat referansı taşır.
-- Ek olarak yaygın temas alerjenleri küratörlükle eklenmiştir.

CREATE TABLE allergen_canonical (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(120) NOT NULL UNIQUE,
    annex_ref  VARCHAR(60),   -- 'Tüzük 1223/2009 Ek III' gibi mevzuat atfı
    source_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bir kanonik alerjenin bilinen yazılışları: Türkçe, İngilizce, INCI.
-- alias tekildir: aynı yazılış iki farklı alerjene işaret edemez, aksi halde
-- eşleştirme belirsizleşir.
CREATE TABLE allergen_alias (
    id           SERIAL PRIMARY KEY,
    canonical_id INT NOT NULL REFERENCES allergen_canonical(id) ON DELETE CASCADE,
    alias        VARCHAR(120) NOT NULL,
    locale       VARCHAR(5),   -- 'tr', 'en', NULL = INCI / uluslararası
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Eşleştirme her zaman küçük harfe indirgenerek yapılır; tekillik de öyle
-- olmalı, yoksa 'Parfüm' ve 'parfüm' iki ayrı satır olarak girebilir.
CREATE UNIQUE INDEX idx_allergen_alias_lower ON allergen_alias (LOWER(alias));
CREATE INDEX idx_allergen_alias_canonical ON allergen_alias (canonical_id);

-- ===== AB Tüzüğü 1223/2009 Ek III — bildirimli koku alerjenleri =====

INSERT INTO allergen_canonical (name, annex_ref, source_url) VALUES
    ('parfüm',                    'Tüzük 1223/2009 Ek III',    'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('linalool',                  'Tüzük 1223/2009 Ek III/84', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('limonen',                   'Tüzük 1223/2009 Ek III/89', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('sitronellol',               'Tüzük 1223/2009 Ek III/87', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('geraniol',                  'Tüzük 1223/2009 Ek III/80', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('sitral',                    'Tüzük 1223/2009 Ek III/70', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('öjenol',                    'Tüzük 1223/2009 Ek III/78', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('kumarin',                   'Tüzük 1223/2009 Ek III/74', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('benzil alkol',              'Tüzük 1223/2009 Ek III/68', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('benzil salisilat',          'Tüzük 1223/2009 Ek III/76', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('benzil benzoat',            'Tüzük 1223/2009 Ek III/75', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('benzil sinamat',            'Tüzük 1223/2009 Ek III/77', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('sinnamal',                  'Tüzük 1223/2009 Ek III/72', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('sinnamil alkol',            'Tüzük 1223/2009 Ek III/71', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('farnesol',                  'Tüzük 1223/2009 Ek III/79', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('hekzil sinnamal',           'Tüzük 1223/2009 Ek III/83', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('hidroksisitronellal',       'Tüzük 1223/2009 Ek III/85', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('izoöjenol',                 'Tüzük 1223/2009 Ek III/86', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('anisil alkol',              'Tüzük 1223/2009 Ek III/69', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('amilsinnamal',              'Tüzük 1223/2009 Ek III/66', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('amilsinnamil alkol',        'Tüzük 1223/2009 Ek III/67', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('metil 2-oktinoat',          'Tüzük 1223/2009 Ek III/90', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('alfa-izometil iyonon',      'Tüzük 1223/2009 Ek III/91', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('butilfenil metilpropional', 'Tüzük 1223/2009 Ek III/82', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('meşe yosunu',               'Tüzük 1223/2009 Ek III/92', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223'),
    ('ağaç yosunu',               'Tüzük 1223/2009 Ek III/93', 'https://eur-lex.europa.eu/legal-content/TR/TXT/?uri=CELEX:32009R1223');

-- ===== Yaygın temas alerjenleri (küratörlü) =====
--
-- 'alkol' ile 'yün alkolü' bilinçli olarak AYRI kanoniklerdir. Yün alkolleri
-- lanolinden gelen sterollerdir; etanolden kaçınan biriyle ilgisi yoktur.

INSERT INTO allergen_canonical (name, annex_ref, source_url) VALUES
    ('lanolin',                NULL, NULL),
    ('yün alkolü',             NULL, NULL),
    ('nikel',                  NULL, NULL),
    ('paraben',                'Tüzük 1223/2009 Ek V', NULL),
    ('formaldehit',            'Tüzük 1223/2009 Ek V', NULL),
    ('sülfat',                 NULL, NULL),
    ('kokamidopropil betain',  NULL, NULL),
    ('hindistan cevizi',       NULL, NULL),
    ('kuruyemiş',              NULL, NULL),
    ('propilen glikol',        NULL, NULL),
    ('alkol',                  NULL, NULL),
    ('bizmut',                 NULL, NULL),
    ('karmin',                 NULL, NULL),
    ('böcek kaynaklı',         NULL, NULL),
    ('salisilat',              NULL, NULL),
    ('retinoid',               NULL, NULL);

-- ===== Yazılış varyantları =====
--
-- Kanonik adın kendisi de bir alias'tır; böylece sorgu tek bir yoldan geçer.

INSERT INTO allergen_alias (canonical_id, alias, locale)
SELECT c.id, v.alias, v.locale
FROM (VALUES
    -- parfüm
    ('parfüm', 'parfüm', 'tr'), ('parfüm', 'parfum', NULL),
    ('parfüm', 'fragrance', 'en'), ('parfüm', 'koku', 'tr'),
    ('parfüm', 'aroma', 'tr'), ('parfüm', 'esans', 'tr'),
    -- Ek III koku alerjenleri
    ('linalool', 'linalool', NULL), ('linalool', 'linalol', 'tr'),
    ('limonen', 'limonen', 'tr'), ('limonen', 'limonene', NULL), ('limonen', 'd-limonene', NULL),
    ('sitronellol', 'sitronellol', 'tr'), ('sitronellol', 'citronellol', NULL),
    ('geraniol', 'geraniol', NULL),
    ('sitral', 'sitral', 'tr'), ('sitral', 'citral', NULL),
    ('öjenol', 'öjenol', 'tr'), ('öjenol', 'eugenol', NULL),
    ('kumarin', 'kumarin', 'tr'), ('kumarin', 'coumarin', NULL),
    ('benzil alkol', 'benzil alkol', 'tr'), ('benzil alkol', 'benzyl alcohol', NULL),
    ('benzil salisilat', 'benzil salisilat', 'tr'), ('benzil salisilat', 'benzyl salicylate', NULL),
    ('benzil benzoat', 'benzil benzoat', 'tr'), ('benzil benzoat', 'benzyl benzoate', NULL),
    ('benzil sinamat', 'benzil sinamat', 'tr'), ('benzil sinamat', 'benzyl cinnamate', NULL),
    ('sinnamal', 'sinnamal', 'tr'), ('sinnamal', 'cinnamal', NULL),
    ('sinnamil alkol', 'sinnamil alkol', 'tr'), ('sinnamil alkol', 'cinnamyl alcohol', NULL),
    ('farnesol', 'farnesol', NULL),
    ('hekzil sinnamal', 'hekzil sinnamal', 'tr'), ('hekzil sinnamal', 'hexyl cinnamal', NULL),
    ('hidroksisitronellal', 'hidroksisitronellal', 'tr'), ('hidroksisitronellal', 'hydroxycitronellal', NULL),
    ('izoöjenol', 'izoöjenol', 'tr'), ('izoöjenol', 'isoeugenol', NULL),
    ('anisil alkol', 'anisil alkol', 'tr'), ('anisil alkol', 'anise alcohol', NULL),
    ('amilsinnamal', 'amilsinnamal', 'tr'), ('amilsinnamal', 'amyl cinnamal', NULL),
    ('amilsinnamil alkol', 'amilsinnamil alkol', 'tr'), ('amilsinnamil alkol', 'amyl cinnamyl alcohol', NULL),
    ('metil 2-oktinoat', 'metil 2-oktinoat', 'tr'), ('metil 2-oktinoat', 'methyl 2-octynoate', NULL),
    ('alfa-izometil iyonon', 'alfa-izometil iyonon', 'tr'), ('alfa-izometil iyonon', 'alpha-isomethyl ionone', NULL),
    ('butilfenil metilpropional', 'butilfenil metilpropional', 'tr'),
    ('butilfenil metilpropional', 'butylphenyl methylpropional', NULL),
    ('butilfenil metilpropional', 'lilial', NULL),
    ('meşe yosunu', 'meşe yosunu', 'tr'), ('meşe yosunu', 'evernia prunastri', NULL),
    ('meşe yosunu', 'oakmoss', 'en'),
    ('ağaç yosunu', 'ağaç yosunu', 'tr'), ('ağaç yosunu', 'evernia furfuracea', NULL),
    ('ağaç yosunu', 'treemoss', 'en'),
    -- Temas alerjenleri
    ('lanolin', 'lanolin', NULL), ('lanolin', 'lanolin alkolü', 'tr'),
    ('yün alkolü', 'yün alkolü', 'tr'), ('yün alkolü', 'wool alcohol', 'en'),
    ('yün alkolü', 'wool alcohols', 'en'),
    ('nikel', 'nikel', 'tr'), ('nikel', 'nickel', 'en'),
    ('paraben', 'paraben', NULL), ('paraben', 'metilparaben', 'tr'),
    ('paraben', 'methylparaben', 'en'), ('paraben', 'propilparaben', 'tr'),
    ('paraben', 'propylparaben', 'en'),
    ('formaldehit', 'formaldehit', 'tr'), ('formaldehit', 'formaldehyde', 'en'),
    ('formaldehit', 'formaldehit salıcı', 'tr'), ('formaldehit', 'dmdm hidantoin', 'tr'),
    ('formaldehit', 'dmdm hydantoin', NULL),
    ('sülfat', 'sülfat', 'tr'), ('sülfat', 'sulfate', 'en'),
    ('sülfat', 'sls', NULL), ('sülfat', 'sodyum lauril sülfat', 'tr'),
    ('sülfat', 'sodium lauryl sulfate', NULL),
    ('kokamidopropil betain', 'kokamidopropil betain', 'tr'),
    ('kokamidopropil betain', 'cocamidopropyl betaine', NULL),
    ('hindistan cevizi', 'hindistan cevizi', 'tr'), ('hindistan cevizi', 'coconut', 'en'),
    ('hindistan cevizi', 'cocos nucifera', NULL),
    ('kuruyemiş', 'kuruyemiş', 'tr'), ('kuruyemiş', 'ağaç kuruyemişi', 'tr'),
    ('kuruyemiş', 'nut', 'en'), ('kuruyemiş', 'tree nut', 'en'),
    ('propilen glikol', 'propilen glikol', 'tr'), ('propilen glikol', 'propylene glycol', NULL),
    ('alkol', 'alkol', 'tr'), ('alkol', 'alcohol', 'en'),
    ('alkol', 'denatüre alkol', 'tr'), ('alkol', 'alcohol denat', NULL),
    ('alkol', 'etanol', 'tr'), ('alkol', 'ethanol', 'en'),
    ('bizmut', 'bizmut', 'tr'), ('bizmut', 'bismuth', 'en'),
    ('karmin', 'karmin', 'tr'), ('karmin', 'carmine', NULL),
    ('karmin', 'koşnil', 'tr'), ('karmin', 'cochineal', 'en'),
    ('böcek kaynaklı', 'böcek kaynaklı', 'tr'), ('böcek kaynaklı', 'insect-derived', 'en'),
    ('salisilat', 'salisilat', 'tr'), ('salisilat', 'salicylate', 'en'),
    ('retinoid', 'retinoid', NULL), ('retinoid', 'retinol', NULL)
) AS v(canonical_name, alias, locale)
JOIN allergen_canonical c ON c.name = v.canonical_name;
