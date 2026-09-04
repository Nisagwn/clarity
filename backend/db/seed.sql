-- Beauty Ingredient Explorer - geliştirme ortamı örnek verisi
--
-- İçerikler ve INCI adları gerçektir. ÜRÜNLER KURGUSALDIR: uydurma marka
-- adları altında makul içerik listeleri, böylece burada gerçekten satın
-- alabileceğiniz hiçbir ürün yanlış tanıtılmaz. Faz 4'te bu dosya Open Beauty
-- Facts'ten gelen gerçek ürün verisiyle değiştirilecek.
--
-- ENDİŞE SEVİYESİ BURADA YAZILMAZ. Puanlar AB Tüzüğü 1223/2009 Eklerinden
-- türetilir; bu dosya yalnızca dayanağı (ingredient_regulatory) kurar.
-- Puanları yazmak için seed'den sonra:
--
--     make score        (veya: cd backend && go run ./cmd/score)
--
-- Tekrar çalıştırmak güvenlidir: her ifade idempotenttir.

-- ===== İçerikler =====
-- name = Türkçe yaygın ad, inci_name = uluslararası INCI adı (çevrilmez).

INSERT INTO ingredients (name, inci_name, description) VALUES
    ('Su',                     'Aqua',                            'Evrensel çözücü; çoğu kozmetik formülün temelini oluşturur.'),
    ('Gliserin',               'Glycerin',                        'Cildin üst katmanlarına su çeken bir nem tutucu.'),
    ('Niasinamid',             'Niacinamide',                     'Aydınlatma ve bariyer desteği için kullanılan B3 vitamini türevi.'),
    ('Hyalüronik Asit',        'Sodium Hyaluronate',              'Kendi ağırlığının katlarca fazlası suyu tutabilen nem tutucu.'),
    ('Skualan',                'Squalane',                        'Hafif, gözenek tıkamayan yumuşatıcı; genelde zeytin veya şeker kamışı kaynaklı.'),
    ('Pantenol',               'Panthenol',                       'Provitamin B5; yatıştırıcı ve nemlendirici.'),
    ('Seramid NP',             'Ceramide NP',                     'Cildin kendi lipidiyle özdeş yapı; nem bariyerini onarmaya yardım eder.'),
    ('Aloe Vera Yaprak Suyu',  'Aloe Barbadensis Leaf Juice',     'Güneş sonrası ve hassas cilt ürünlerinde kullanılan yatıştırıcı bitki özütü.'),
    ('Tokoferol (E Vitamini)', 'Tocopherol',                      'Antioksidan; aynı zamanda formüldeki yağların bozulmasını geciktirir.'),
    ('Shea Yağı',              'Butyrospermum Parkii Butter',     'Kuru cildi besleyen, kuruyemiş kaynaklı zengin bir yağ.'),
    ('Hindistan Cevizi Yağı',  'Cocos Nucifera Oil',              'Nem hapseden bitkisel yağ; bazı cilt tiplerinde gözenek tıkayıcıdır.'),
    ('Titanyum Dioksit',       'Titanium Dioxide (CI 77891)',     'Mineral UV filtresi ve beyaz pigment.'),
    ('Çinko Oksit',            'Zinc Oxide',                      'Geniş spektrumlu mineral UV filtresi; hafif yatıştırıcı etkisi de vardır.'),
    ('Demir Oksitler',         'Iron Oxides (CI 77491)',          'Kırmızı, sarı ve siyah tonları veren mineral pigmentler.'),
    ('Mika',                   'Mica (CI 77019)',                 'Işıltı ve yumuşak odak etkisi için kullanılan mineral.'),
    ('Silika',                 'Silica',                          'Matlaştırıcı, yağ emici mineral toz.'),
    ('Talk',                   'Talc',                            'Yumuşak mineral dolgu; asbest riski nedeniyle tedarik kaynağı önemlidir.'),
    ('Bizmut Oksiklorür',      'Bismuth Oxychloride',             'Sedefli mineral pigment; hassas ciltte tahrişe yol açabilir.'),
    ('Karmin',                 'Carmine (CI 75470)',              'Koşnil böceğinden elde edilen kırmızı pigment.'),
    ('Dimetikon',              'Dimethicone',                     'Cilt dokusunu pürüzsüzleştiren ve ince çizgileri gizleyen silikon.'),
    ('Siklopentasiloksan',     'Cyclopentasiloxane',              'İpeksi bir kayganlık veren ve uçan silikon.'),
    ('Setearil Alkol',         'Cetearyl Alcohol',                'Yumuşatıcı ve emülgatör olarak kullanılan yağ alkolü; kurutucu değildir.'),
    ('Poligliseril-3 Diizostearat', 'Polyglyceryl-3 Diisostearate', 'Su-yağ formülleri için yumuşak, noniyonik emülgatör.'),
    ('Fenoksietanol',          'Phenoxyethanol',                  'Geniş spektrumlu koruyucu; AB''de %1 ile sınırlıdır.'),
    ('Metilparaben',           'Methylparaben',                   'Paraben grubu koruyucu; etkilidir ama tüketiciler yaygın olarak kaçınır.'),
    ('DMDM Hidantoin',         'DMDM Hydantoin',                  'Yavaşça formaldehit salarak koruyuculuk sağlayan madde.'),
    ('Propilen Glikol',        'Propylene Glycol',                'Nem tutucu ve çözücü; bazı kişilerde bilinen bir temas alerjenidir.'),
    ('Denatüre Alkol',         'Alcohol Denat.',                  'Hızlı uçan çözücü; yüksek oranlarda hassasiyet yaratabilir.'),
    ('Sodyum Lauril Sülfat',   'Sodium Lauryl Sulfate',           'Cilt bariyerini sıyırabilen güçlü anyonik yüzey aktif madde.'),
    ('Kokamidopropil Betain',  'Cocamidopropyl Betaine',          'Hindistan cevizi kaynaklı daha yumuşak yüzey aktif; tanınmış temas alerjeni.'),
    ('Salisilik Asit',         'Salicylic Acid',                  'Gözenek içinde peeling yapan beta hidroksi asit.'),
    ('Retinol',                'Retinol',                         'Hücre yenilenmesi için A vitamini türevi; gebelikte kullanılmaz.'),
    ('Lanolin',                'Lanolin',                         'Yün kaynaklı nem hapsedici; klasik bir temas alerjeni.'),
    ('Parfüm',                 'Parfum (Fragrance)',              'İçeriği açıklanmayan koku karışımı; en yaygın kozmetik alerjeni.'),
    ('Linalool',               'Linalool',                        'Okside olduğunda hassasiyet yaratan koku bileşeni.'),
    ('Limonen',                'Limonene',                        'Narenciye kaynaklı koku bileşeni; AB''nin 26 bildirimli alerjeninden biri.'),
    ('Benzil Salisilat',       'Benzyl Salicylate',               'AB bildirimli alerjen listesindeki koku maddesi.')
ON CONFLICT (name) DO NOTHING;

-- ===== Mevzuat kayıtları =====
--
-- Puanın dayanağı. 000005_regulatory_scoring göçü bu satırları zaten
-- yazıyor; ancak göç boş bir veritabanında çalıştığında eşleşecek içerik
-- olmadığı için hiçbir şey yazamaz. Bu yüzden seed de aynı veriyi kurar.
--
-- Yalnızca TANIMI GEREĞİ kesin olan kayıtlar var: Ek III bildirimli koku
-- alerjenleri, Ek V koruyucular, Ek VI UV filtreleri, Ek IV renklendiriciler.
-- Geri kalan içerikler BİLİNÇLİ olarak boş: tahmin edilmiş bir Ek numarası,
-- tahmin edilmiş bir puandan kötüdür: mevzuata atıf yapıyormuş gibi görünüp
-- yanlış olur. Tam CosIng dışa aktarımı için: backend/scripts/import-cosing

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
JOIN ingredients i ON i.name = v.ing_name
ON CONFLICT (ingredient_id) DO UPDATE SET
    annex               = EXCLUDED.annex,
    annex_entry         = EXCLUDED.annex_entry,
    declarable_allergen = EXCLUDED.declarable_allergen,
    restriction         = EXCLUDED.restriction,
    source_url          = EXCLUDED.source_url,
    fetched_at          = CURRENT_TIMESTAMP;

-- ===== Alerjen eşleştirmeleri =====

INSERT INTO ingredient_allergens (ingredient_id, allergen_name, severity)
SELECT i.id, v.allergen, v.severity
FROM (VALUES
    ('Parfüm',                  'parfüm',                 8),
    ('Linalool',                'parfüm',                 6),
    ('Linalool',                'linalool',               6),
    ('Limonen',                 'parfüm',                 6),
    ('Limonen',                 'limonen',                6),
    ('Benzil Salisilat',        'parfüm',                 6),
    ('Benzil Salisilat',        'salisilat',              5),
    ('Salisilik Asit',          'salisilat',              5),
    ('Lanolin',                 'lanolin',                6),
    ('Lanolin',                 'yün alkolü',             6),
    ('Karmin',                  'karmin',                 5),
    ('Karmin',                  'böcek kaynaklı',         5),
    ('Metilparaben',            'paraben',                5),
    ('DMDM Hidantoin',          'formaldehit',            9),
    ('Sodyum Lauril Sülfat',    'sülfat',                 5),
    ('Kokamidopropil Betain',   'kokamidopropil betain',  5),
    ('Kokamidopropil Betain',   'hindistan cevizi',       4),
    ('Hindistan Cevizi Yağı',   'hindistan cevizi',       4),
    ('Shea Yağı',               'kuruyemiş',              3),
    ('Propilen Glikol',         'propilen glikol',        4),
    ('Denatüre Alkol',          'alkol',                  4),
    ('Bizmut Oksiklorür',       'bizmut',                 4),
    -- Mineral pigmentler ve dolgular eser miktarda nikel taşıyabilir; nikel
    -- temas alerjisi olanlar için bu önemlidir.
    ('Talk',                    'nikel',                  4),
    ('Mika',                    'nikel',                  4),
    ('Demir Oksitler',          'nikel',                  3),
    ('Retinol',                 'retinoid',               6)
) AS v(ing_name, allergen, severity)
JOIN ingredients i ON i.name = v.ing_name
ON CONFLICT (ingredient_id, allergen_name) DO NOTHING;

-- ===== Faydalar =====

INSERT INTO ingredient_benefits (ingredient_id, benefit)
SELECT i.id, v.benefit
FROM (VALUES
    ('Su',                      'çözücü'),
    ('Gliserin',                'nemlendirme'),
    ('Gliserin',                'nem tutucu'),
    ('Niasinamid',              'aydınlatma'),
    ('Niasinamid',              'gözenek sıkılaştırma'),
    ('Niasinamid',              'bariyer desteği'),
    ('Hyalüronik Asit',         'nemlendirme'),
    ('Hyalüronik Asit',         'dolgunlaştırma'),
    ('Skualan',                 'nemlendirme'),
    ('Skualan',                 'gözenek tıkamaz'),
    ('Pantenol',                'yatıştırma'),
    ('Pantenol',                'nemlendirme'),
    ('Seramid NP',              'bariyer onarımı'),
    ('Aloe Vera Yaprak Suyu',   'yatıştırma'),
    ('Tokoferol (E Vitamini)',  'antioksidan'),
    ('Shea Yağı',               'besleyici'),
    ('Shea Yağı',               'bariyer onarımı'),
    ('Hindistan Cevizi Yağı',   'nem hapsedici'),
    ('Titanyum Dioksit',        'UV koruması'),
    ('Titanyum Dioksit',        'beyazlatma'),
    ('Çinko Oksit',             'UV koruması'),
    ('Çinko Oksit',             'yatıştırma'),
    ('Demir Oksitler',          'pigment'),
    ('Mika',                    'ışık dağıtma'),
    ('Silika',                  'yağ emme'),
    ('Silika',                  'matlaştırma'),
    ('Dimetikon',               'pürüzsüzleştirme'),
    ('Dimetikon',               'gözenek gizleme'),
    ('Siklopentasiloksan',      'ipeksi uygulama'),
    ('Setearil Alkol',          'yumuşatıcı'),
    ('Fenoksietanol',           'koruyucu'),
    ('Metilparaben',            'koruyucu'),
    ('DMDM Hidantoin',          'koruyucu'),
    ('Salisilik Asit',          'peeling'),
    ('Salisilik Asit',          'akne kontrolü'),
    ('Retinol',                 'hücre yenilenmesi'),
    ('Retinol',                 'yaşlanma karşıtı'),
    ('Lanolin',                 'nem hapsedici'),
    ('Kokamidopropil Betain',   'nazik temizlik'),
    ('Sodyum Lauril Sülfat',    'derin temizlik')
) AS v(ing_name, benefit)
JOIN ingredients i ON i.name = v.ing_name
ON CONFLICT DO NOTHING;

-- ===== Cilt tipi uygunluğu =====
-- Değerler API sözleşmesinin parçası olduğu için İngilizce kalır; arayüz
-- bunları Türkçe etiketlerle gösterir. 'all' joker değerdir.

INSERT INTO ingredient_skin_types (ingredient_id, skin_type)
SELECT i.id, v.skin_type
FROM (VALUES
    ('Su',                      'all'),
    ('Gliserin',                'all'),
    ('Hyalüronik Asit',         'all'),
    ('Skualan',                 'all'),
    ('Pantenol',                'all'),
    ('Tokoferol (E Vitamini)',  'all'),
    ('Niasinamid',              'oily'),
    ('Niasinamid',              'combination'),
    ('Niasinamid',              'sensitive'),
    ('Niasinamid',              'normal'),
    ('Seramid NP',              'dry'),
    ('Seramid NP',              'sensitive'),
    ('Seramid NP',              'normal'),
    ('Aloe Vera Yaprak Suyu',   'sensitive'),
    ('Aloe Vera Yaprak Suyu',   'oily'),
    ('Shea Yağı',               'dry'),
    ('Shea Yağı',               'normal'),
    ('Hindistan Cevizi Yağı',   'dry'),
    ('Lanolin',                 'dry'),
    ('Titanyum Dioksit',        'oily'),
    ('Titanyum Dioksit',        'combination'),
    ('Titanyum Dioksit',        'sensitive'),
    ('Çinko Oksit',             'oily'),
    ('Çinko Oksit',             'sensitive'),
    ('Silika',                  'oily'),
    ('Silika',                  'combination'),
    ('Talk',                    'oily'),
    ('Dimetikon',               'dry'),
    ('Dimetikon',               'normal'),
    ('Dimetikon',               'combination'),
    ('Setearil Alkol',          'dry'),
    ('Setearil Alkol',          'normal'),
    ('Salisilik Asit',          'oily'),
    ('Salisilik Asit',          'combination'),
    ('Retinol',                 'normal'),
    ('Retinol',                 'oily'),
    ('Retinol',                 'combination'),
    ('Denatüre Alkol',          'oily'),
    ('Sodyum Lauril Sülfat',    'oily'),
    ('Kokamidopropil Betain',   'normal'),
    ('Mika',                    'all'),
    ('Demir Oksitler',          'all')
) AS v(ing_name, skin_type)
JOIN ingredients i ON i.name = v.ing_name
ON CONFLICT (ingredient_id, skin_type) DO NOTHING;

-- ===== Ürünler (kurgusal markalar, yalnızca geliştirme için) =====

INSERT INTO products (name, brand, gtin, price, currency, category, description) VALUES
    ('Işıltılı İpek Fondöten SPF 20',    'Maison Lumiere', '8690000000011', 52.00, 'USD', 'fondöten',      'Mineral güneş koruyuculu, aydınlık bitişli orta kapatıcılıkta fondöten.'),
    ('Günlük Işıltı Fondöten SPF 15',    'Rue Belle',      '8690000000028', 18.50, 'USD', 'fondöten',      'Doğal bitişli, hafif günlük fondöten.'),
    ('Kadife Mat Fondöten',              'Atelier Noir',   '8690000000035', 46.00, 'USD', 'fondöten',      'Yağlı ciltler için uzun süre kalıcı mat fondöten.'),
    ('İkinci Ten Renkli Serum',          'Rue Belle',      '8690000000042', 21.00, 'USD', 'fondöten',      'Niasinamid içeren ince örtücülükte renkli serum.'),
    ('Yumuşak Odak Sabitleyici Pudra',   'Maison Lumiere', '8690000000059', 38.00, 'USD', 'pudra',         'Gözenekleri gizleyen ince öğütülmüş pudra.'),
    ('Gözenek Gizleyici Toz Pudra',      'Rue Belle',      '8690000000066', 12.99, 'USD', 'pudra',         'Parlama kontrolü için şeffaf toz pudra.'),
    ('Gül Yaprağı Ruj',                  'Atelier Noir',   '8690000000073', 34.00, 'USD', 'ruj',           'Sıcak gül tonunda saten bitişli ruj.'),
    ('Kremsi Nude Ruj',                  'Rue Belle',      '8690000000080', 9.99,  'USD', 'ruj',           'Her güne uygun kremsi nude ruj.'),
    ('Bariyer Onarıcı Temizleyici',      'Derma Basics',   '8690000000097', 14.00, 'USD', 'temizleyici',   'Seramid içeren köpürmeyen temizleyici.'),
    ('Nazik Köpük Temizleyici',          'Rue Belle',      '8690000000103', 8.50,  'USD', 'temizleyici',   'Günlük kullanıma uygun köpüren yüz temizleyici.'),
    ('Gece Retinol Kremi',               'Derma Basics',   '8690000000110', 42.00, 'USD', 'nemlendirici',  '%0,3 retinol içeren gece kremi.'),
    ('Sakinleştirici Seramid Nemlendirici','Derma Basics', '8690000000127', 26.00, 'USD', 'nemlendirici',  'Hassas ciltler için parfümsüz nemlendirici.'),
    ('Uzun Süre Kalıcı Likit Eyeliner',  'Atelier Noir',   '8690000000134', 24.00, 'USD', 'eyeliner',      'Suya dayanıklı keçe uçlu likit eyeliner.'),
    ('Arındırıcı Akne Serumu',           'Derma Basics',   '8690000000141', 19.00, 'USD', 'serum',         'Sıkışmış ciltler için %2 salisilik asit serumu.')
ON CONFLICT (gtin, brand) DO NOTHING;

-- ===== Ürün içerik listeleri (order_index = INCI listesindeki sıra) =====

INSERT INTO product_ingredients (product_id, ingredient_id, order_index)
SELECT p.id, i.id, v.ord
FROM (VALUES
    -- Işıltılı İpek Fondöten SPF 20
    ('Işıltılı İpek Fondöten SPF 20', 'Su', 1),
    ('Işıltılı İpek Fondöten SPF 20', 'Siklopentasiloksan', 2),
    ('Işıltılı İpek Fondöten SPF 20', 'Titanyum Dioksit', 3),
    ('Işıltılı İpek Fondöten SPF 20', 'Gliserin', 4),
    ('Işıltılı İpek Fondöten SPF 20', 'Dimetikon', 5),
    ('Işıltılı İpek Fondöten SPF 20', 'Mika', 6),
    ('Işıltılı İpek Fondöten SPF 20', 'Demir Oksitler', 7),
    ('Işıltılı İpek Fondöten SPF 20', 'Skualan', 8),
    ('Işıltılı İpek Fondöten SPF 20', 'Tokoferol (E Vitamini)', 9),
    ('Işıltılı İpek Fondöten SPF 20', 'Fenoksietanol', 10),
    ('Işıltılı İpek Fondöten SPF 20', 'Parfüm', 11),
    ('Işıltılı İpek Fondöten SPF 20', 'Linalool', 12),

    -- Günlük Işıltı Fondöten SPF 15 (yukarıdakinin parfümsüz muadili)
    ('Günlük Işıltı Fondöten SPF 15', 'Su', 1),
    ('Günlük Işıltı Fondöten SPF 15', 'Siklopentasiloksan', 2),
    ('Günlük Işıltı Fondöten SPF 15', 'Titanyum Dioksit', 3),
    ('Günlük Işıltı Fondöten SPF 15', 'Gliserin', 4),
    ('Günlük Işıltı Fondöten SPF 15', 'Dimetikon', 5),
    ('Günlük Işıltı Fondöten SPF 15', 'Mika', 6),
    ('Günlük Işıltı Fondöten SPF 15', 'Demir Oksitler', 7),
    ('Günlük Işıltı Fondöten SPF 15', 'Skualan', 8),
    ('Günlük Işıltı Fondöten SPF 15', 'Tokoferol (E Vitamini)', 9),
    ('Günlük Işıltı Fondöten SPF 15', 'Fenoksietanol', 10),

    -- Kadife Mat Fondöten
    ('Kadife Mat Fondöten', 'Su', 1),
    ('Kadife Mat Fondöten', 'Dimetikon', 2),
    ('Kadife Mat Fondöten', 'Talk', 3),
    ('Kadife Mat Fondöten', 'Demir Oksitler', 4),
    ('Kadife Mat Fondöten', 'Mika', 5),
    ('Kadife Mat Fondöten', 'Bizmut Oksiklorür', 6),
    ('Kadife Mat Fondöten', 'Denatüre Alkol', 7),
    ('Kadife Mat Fondöten', 'Fenoksietanol', 8),
    ('Kadife Mat Fondöten', 'Parfüm', 9),
    ('Kadife Mat Fondöten', 'Limonen', 10),

    -- İkinci Ten Renkli Serum
    ('İkinci Ten Renkli Serum', 'Su', 1),
    ('İkinci Ten Renkli Serum', 'Gliserin', 2),
    ('İkinci Ten Renkli Serum', 'Niasinamid', 3),
    ('İkinci Ten Renkli Serum', 'Hyalüronik Asit', 4),
    ('İkinci Ten Renkli Serum', 'Titanyum Dioksit', 5),
    ('İkinci Ten Renkli Serum', 'Demir Oksitler', 6),
    ('İkinci Ten Renkli Serum', 'Skualan', 7),
    ('İkinci Ten Renkli Serum', 'Pantenol', 8),
    ('İkinci Ten Renkli Serum', 'Fenoksietanol', 9),

    -- Yumuşak Odak Sabitleyici Pudra
    ('Yumuşak Odak Sabitleyici Pudra', 'Talk', 1),
    ('Yumuşak Odak Sabitleyici Pudra', 'Mika', 2),
    ('Yumuşak Odak Sabitleyici Pudra', 'Silika', 3),
    ('Yumuşak Odak Sabitleyici Pudra', 'Titanyum Dioksit', 4),
    ('Yumuşak Odak Sabitleyici Pudra', 'Demir Oksitler', 5),
    ('Yumuşak Odak Sabitleyici Pudra', 'Tokoferol (E Vitamini)', 6),
    ('Yumuşak Odak Sabitleyici Pudra', 'Parfüm', 7),

    -- Gözenek Gizleyici Toz Pudra (sabitleyici pudranın muadili)
    ('Gözenek Gizleyici Toz Pudra', 'Talk', 1),
    ('Gözenek Gizleyici Toz Pudra', 'Mika', 2),
    ('Gözenek Gizleyici Toz Pudra', 'Silika', 3),
    ('Gözenek Gizleyici Toz Pudra', 'Titanyum Dioksit', 4),
    ('Gözenek Gizleyici Toz Pudra', 'Demir Oksitler', 5),
    ('Gözenek Gizleyici Toz Pudra', 'Tokoferol (E Vitamini)', 6),

    -- Gül Yaprağı Ruj
    ('Gül Yaprağı Ruj', 'Shea Yağı', 1),
    ('Gül Yaprağı Ruj', 'Hindistan Cevizi Yağı', 2),
    ('Gül Yaprağı Ruj', 'Lanolin', 3),
    ('Gül Yaprağı Ruj', 'Karmin', 4),
    ('Gül Yaprağı Ruj', 'Mika', 5),
    ('Gül Yaprağı Ruj', 'Demir Oksitler', 6),
    ('Gül Yaprağı Ruj', 'Tokoferol (E Vitamini)', 7),
    ('Gül Yaprağı Ruj', 'Parfüm', 8),
    ('Gül Yaprağı Ruj', 'Limonen', 9),

    -- Kremsi Nude Ruj (lanolinsiz ve parfümsüz muadil)
    ('Kremsi Nude Ruj', 'Shea Yağı', 1),
    ('Kremsi Nude Ruj', 'Hindistan Cevizi Yağı', 2),
    ('Kremsi Nude Ruj', 'Karmin', 3),
    ('Kremsi Nude Ruj', 'Mika', 4),
    ('Kremsi Nude Ruj', 'Demir Oksitler', 5),
    ('Kremsi Nude Ruj', 'Tokoferol (E Vitamini)', 6),
    ('Kremsi Nude Ruj', 'Skualan', 7),

    -- Bariyer Onarıcı Temizleyici
    ('Bariyer Onarıcı Temizleyici', 'Su', 1),
    ('Bariyer Onarıcı Temizleyici', 'Gliserin', 2),
    ('Bariyer Onarıcı Temizleyici', 'Setearil Alkol', 3),
    ('Bariyer Onarıcı Temizleyici', 'Seramid NP', 4),
    ('Bariyer Onarıcı Temizleyici', 'Pantenol', 5),
    ('Bariyer Onarıcı Temizleyici', 'Aloe Vera Yaprak Suyu', 6),
    ('Bariyer Onarıcı Temizleyici', 'Fenoksietanol', 7),

    -- Nazik Köpük Temizleyici
    ('Nazik Köpük Temizleyici', 'Su', 1),
    ('Nazik Köpük Temizleyici', 'Gliserin', 2),
    ('Nazik Köpük Temizleyici', 'Sodyum Lauril Sülfat', 3),
    ('Nazik Köpük Temizleyici', 'Kokamidopropil Betain', 4),
    ('Nazik Köpük Temizleyici', 'Aloe Vera Yaprak Suyu', 5),
    ('Nazik Köpük Temizleyici', 'Fenoksietanol', 6),
    ('Nazik Köpük Temizleyici', 'Parfüm', 7),

    -- Gece Retinol Kremi
    ('Gece Retinol Kremi', 'Su', 1),
    ('Gece Retinol Kremi', 'Gliserin', 2),
    ('Gece Retinol Kremi', 'Retinol', 3),
    ('Gece Retinol Kremi', 'Skualan', 4),
    ('Gece Retinol Kremi', 'Tokoferol (E Vitamini)', 5),
    ('Gece Retinol Kremi', 'Setearil Alkol', 6),
    ('Gece Retinol Kremi', 'Dimetikon', 7),
    ('Gece Retinol Kremi', 'Fenoksietanol', 8),

    -- Sakinleştirici Seramid Nemlendirici
    ('Sakinleştirici Seramid Nemlendirici', 'Su', 1),
    ('Sakinleştirici Seramid Nemlendirici', 'Gliserin', 2),
    ('Sakinleştirici Seramid Nemlendirici', 'Seramid NP', 3),
    ('Sakinleştirici Seramid Nemlendirici', 'Skualan', 4),
    ('Sakinleştirici Seramid Nemlendirici', 'Pantenol', 5),
    ('Sakinleştirici Seramid Nemlendirici', 'Setearil Alkol', 6),
    ('Sakinleştirici Seramid Nemlendirici', 'Niasinamid', 7),
    ('Sakinleştirici Seramid Nemlendirici', 'Fenoksietanol', 8),

    -- Uzun Süre Kalıcı Likit Eyeliner (formaldehit salıcı; alerjen demosu için)
    ('Uzun Süre Kalıcı Likit Eyeliner', 'Su', 1),
    ('Uzun Süre Kalıcı Likit Eyeliner', 'Demir Oksitler', 2),
    ('Uzun Süre Kalıcı Likit Eyeliner', 'Propilen Glikol', 3),
    ('Uzun Süre Kalıcı Likit Eyeliner', 'DMDM Hidantoin', 4),
    ('Uzun Süre Kalıcı Likit Eyeliner', 'Denatüre Alkol', 5),
    ('Uzun Süre Kalıcı Likit Eyeliner', 'Fenoksietanol', 6),

    -- Arındırıcı Akne Serumu
    ('Arındırıcı Akne Serumu', 'Su', 1),
    ('Arındırıcı Akne Serumu', 'Salisilik Asit', 2),
    ('Arındırıcı Akne Serumu', 'Niasinamid', 3),
    ('Arındırıcı Akne Serumu', 'Gliserin', 4),
    ('Arındırıcı Akne Serumu', 'Propilen Glikol', 5),
    ('Arındırıcı Akne Serumu', 'Aloe Vera Yaprak Suyu', 6),
    ('Arındırıcı Akne Serumu', 'Fenoksietanol', 7)
) AS v(product_name, ing_name, ord)
JOIN products p ON p.name = v.product_name
JOIN ingredients i ON i.name = v.ing_name
ON CONFLICT (product_id, ingredient_id) DO NOTHING;

-- ===== Örnek kullanıcı profili =====

INSERT INTO user_profiles (email, skin_type) VALUES
    ('demo@beauty.local', 'sensitive')
ON CONFLICT (email) DO NOTHING;

INSERT INTO user_allergens (user_id, allergen_name, severity)
SELECT u.id, v.allergen, v.severity
FROM (VALUES
    ('demo@beauty.local', 'parfüm',      8),
    ('demo@beauty.local', 'nikel',       6),
    ('demo@beauty.local', 'formaldehit', 9)
) AS v(email, allergen, severity)
JOIN user_profiles u ON u.email = v.email
ON CONFLICT (user_id, allergen_name) DO NOTHING;
