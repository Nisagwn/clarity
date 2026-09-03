-- Kimlik doğrulama ve rıza kaydı
--
-- Neden: /profiles/:id çıplak bir tam sayı alıyordu; sayıyı artıran herkes
-- başkasının alerjen listesini okuyabiliyordu. Alerjen listesi hem KVKK m.6'da
-- "özel nitelikli kişisel veri" hem GDPR m.9'da "special category data" —
-- yani sağlık verisi.
--
-- KVKK'nın sağlık verisi için rızasız işleme istisnaları yalnızca sır saklama
-- yükümlüsü kişiler ve yetkili sağlık kuruluşları için geçerli; Clarity
-- bunların dışında. Bu yüzden açık rıza zorunlu ve kanıtlanabilir olmalı.

ALTER TABLE user_profiles
    ADD COLUMN password_hash     TEXT,
    ADD COLUMN email_verified_at TIMESTAMP,
    ADD COLUMN deleted_at        TIMESTAMP;

-- E-posta artık kimlik: hesap oluşturmak için zorunlu.
-- Mevcut satırlar (örnek veri) NULL şifreyle kalır ve giriş yapamaz.
CREATE UNIQUE INDEX idx_user_profiles_email_active
    ON user_profiles (LOWER(email)) WHERE deleted_at IS NULL;

-- Rızanın kanıtlanması KVKK ve GDPR'da veri sorumlusunun yükümlülüğü.
-- Hangi metin sürümüne, ne zaman, hangi bağlamda rıza verildiği tutulur.
-- Geri alma da bir satırdır (granted = false): rıza geçmişi silinmez.
CREATE TABLE consent_log (
    id             SERIAL PRIMARY KEY,
    user_id        INT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    consent_type   VARCHAR(50) NOT NULL,   -- 'health_data' | 'marketing'
    granted        BOOLEAN NOT NULL,
    policy_version VARCHAR(20) NOT NULL,
    ip_address     INET,
    user_agent     TEXT,
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_consent_log_user ON consent_log (user_id, consent_type, created_at DESC);

-- Şifre sıfırlama jetonları. Jetonun kendisi değil, SHA-256 özeti saklanır:
-- veritabanı sızarsa jetonlar kullanılamaz olmalı.
CREATE TABLE password_reset_token (
    id         SERIAL PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    used_at    TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_password_reset_user ON password_reset_token (user_id);
