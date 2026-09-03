package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// Bu testler Faz 2'nin varlık sebebini koruyor: alerjen listesi hem KVKK'da
// hem GDPR'da özel nitelikli sağlık verisi. Önceki tasarımda /profiles/:id
// çıplak bir tam sayı alıyordu ve sayıyı artıran herkes başkasının listesini
// okuyabiliyordu.

func init() {
	// Testler kendi imzalama anahtarını kullanır.
	if os.Getenv("JWT_SECRET") == "" {
		os.Setenv("JWT_SECRET", "test-imzalama-anahtari-yalnizca-testler-icin")
	}
}

// session, bir isteğe oturum cookie'sini taşır.
type session struct {
	cookie *http.Cookie
}

// doAuthed, oturumlu bir istek yapar.
func doAuthed(t *testing.T, r *gin.Engine, s *session, method, path string, body any, out any) int {
	t.Helper()

	var reader *strings.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("gövde kodlanamadı: %v", err)
		}
		reader = strings.NewReader(string(raw))
	} else {
		reader = strings.NewReader("")
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if s != nil && s.cookie != nil {
		req.AddCookie(s.cookie)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if out != nil && w.Body.Len() > 0 {
		_ = json.Unmarshal(w.Body.Bytes(), out)
	}
	return w.Code
}

// registerUser, bir hesap açıp oturumunu döndürür.
func registerUser(t *testing.T, r *gin.Engine, email string, allergens []string, healthConsent bool) (*session, int) {
	t.Helper()

	body := map[string]any{
		"email":               email,
		"password":            "cok-guclu-bir-parola",
		"skin_type":           "sensitive",
		"health_data_consent": healthConsent,
		"marketing_consent":   false,
		"allergens":           allergens,
	}

	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("kayıt başarısız (%d): %s", w.Code, w.Body.String())
	}

	var resp struct {
		ID int `json:"id"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	for _, ck := range w.Result().Cookies() {
		if ck.Name == "clarity_session" {
			return &session{cookie: ck}, resp.ID
		}
	}
	t.Fatal("kayıt sonrası oturum cookie'si gelmedi")
	return nil, 0
}

// TestProfileRequiresAuth, kimliksiz isteğin profile erişemediğini doğrular.
func TestProfileRequiresAuth(t *testing.T) {
	_, r, _ := newTestServer(t)

	code := doAuthed(t, r, nil, http.MethodGet, "/profiles/me", nil, nil)
	mustStatus(t, code, http.StatusUnauthorized, "kimliksiz profil okuma")

	code = doAuthed(t, r, nil, http.MethodPut, "/profiles/me", map[string]any{"skin_type": "dry"}, nil)
	mustStatus(t, code, http.StatusUnauthorized, "kimliksiz profil yazma")

	code = doAuthed(t, r, nil, http.MethodDelete, "/auth/account", nil, nil)
	mustStatus(t, code, http.StatusUnauthorized, "kimliksiz hesap silme")
}

// TestProfileIsolation, bir kullanıcının başkasının verisini göremediğini
// doğrular. Eski /profiles/:id tasarımında bu mümkündü.
func TestProfileIsolation(t *testing.T) {
	_, r, _ := newTestServer(t)

	ayseSession, _ := registerUser(t, r, "ayse@ornek.com", []string{"parfüm"}, true)
	_, _ = registerUser(t, r, "mehmet@ornek.com", []string{"nikel"}, true)

	// Ayşe kendi profilini okuyor: yalnızca kendi alerjenini görmeli.
	var profile struct {
		ID        int      `json:"id"`
		Email     string   `json:"email"`
		Allergens []string `json:"allergens"`
	}
	code := doAuthed(t, r, ayseSession, http.MethodGet, "/profiles/me", nil, &profile)
	mustStatus(t, code, statusOK, "kendi profilini okuma")

	if profile.Email != "ayse@ornek.com" {
		t.Errorf("yanlış profil döndü: %s", profile.Email)
	}
	if !contains(profile.Allergens, "parfüm") {
		t.Errorf("kendi alerjeni eksik: %v", profile.Allergens)
	}
	if contains(profile.Allergens, "nikel") {
		t.Errorf("BAŞKASININ alerjeni sızdı: %v", profile.Allergens)
	}
}

// TestHealthDataRequiresConsent, açık rıza olmadan alerjen kaydedilemediğini
// doğrular. KVKK m.6 ve GDPR m.9 gereği.
func TestHealthDataRequiresConsent(t *testing.T) {
	_, r, _ := newTestServer(t)

	// Rıza yok ama alerjen gönderiliyor: reddedilmeli.
	body := map[string]any{
		"email":               "rizasiz@ornek.com",
		"password":            "cok-guclu-bir-parola",
		"health_data_consent": false,
		"allergens":           []string{"parfüm"},
	}
	code := doJSON(t, r, http.MethodPost, "/auth/register", body, nil)
	mustStatus(t, code, http.StatusBadRequest, "rızasız alerjen kaydı")
}

// TestConsentWithdrawalDeletesData, rıza geri alındığında alerjen verisinin
// derhal silindiğini doğrular. Rızanın geri alınabilir olması yetmez;
// geri alınınca işlemenin durması gerekir.
func TestConsentWithdrawalDeletesData(t *testing.T) {
	_, r, db := newTestServer(t)

	sess, userID := registerUser(t, r, "geri@ornek.com", []string{"parfüm", "nikel"}, true)

	var before int
	if err := db.QueryRow("SELECT count(*) FROM user_allergens WHERE user_id = $1", userID).Scan(&before); err != nil {
		t.Fatalf("sayım başarısız: %v", err)
	}
	if before != 2 {
		t.Fatalf("kurulum hatalı: %d alerjen bekleniyordu, %d var", 2, before)
	}

	code := doAuthed(t, r, sess, http.MethodPost, "/auth/consent", map[string]any{
		"consent_type": "health_data",
		"granted":      false,
	}, nil)
	mustStatus(t, code, statusOK, "rıza geri alma")

	var after int
	if err := db.QueryRow("SELECT count(*) FROM user_allergens WHERE user_id = $1", userID).Scan(&after); err != nil {
		t.Fatalf("sayım başarısız: %v", err)
	}
	if after != 0 {
		t.Errorf("rıza geri alındı ama %d alerjen kaydı duruyor", after)
	}

	// Rıza yokken yeniden alerjen yazmak da engellenmeli.
	code = doAuthed(t, r, sess, http.MethodPut, "/profiles/me", map[string]any{
		"skin_type": "dry",
		"allergens": []string{"lanolin"},
	}, nil)
	mustStatus(t, code, http.StatusForbidden, "rızasız alerjen yazma")
}

// TestConsentIsLogged, rızanın kanıtlanabilir biçimde kaydedildiğini doğrular.
// Rızanın ispatı veri sorumlusunun yükümlülüğü.
func TestConsentIsLogged(t *testing.T) {
	_, r, db := newTestServer(t)

	_, userID := registerUser(t, r, "kayit@ornek.com", []string{"parfüm"}, true)

	rows, err := db.Query(
		`SELECT consent_type, granted, policy_version FROM consent_log
		 WHERE user_id = $1 ORDER BY consent_type`, userID)
	if err != nil {
		t.Fatalf("rıza kaydı okunamadı: %v", err)
	}
	defer rows.Close()

	seen := map[string]bool{}
	for rows.Next() {
		var ctype, version string
		var granted bool
		if err := rows.Scan(&ctype, &granted, &version); err != nil {
			t.Fatalf("satır okunamadı: %v", err)
		}
		seen[ctype] = granted
		if version == "" {
			t.Errorf("%s rızasında metin sürümü kayıtlı değil", ctype)
		}
	}

	// "Hayır" da kaydedilmeli: rıza verilmediğinin kanıtı da gerekiyor.
	if _, ok := seen["marketing"]; !ok {
		t.Error("reddedilen pazarlama rızası kaydedilmemiş")
	}
	if !seen["health_data"] {
		t.Error("sağlık verisi rızası olumlu kaydedilmemiş")
	}
}

// TestDeleteAccountRemovesData, hesap silmenin ilişkili veriyi de sildiğini
// doğrular (silme hakkı).
func TestDeleteAccountRemovesData(t *testing.T) {
	_, r, db := newTestServer(t)

	sess, userID := registerUser(t, r, "silinecek@ornek.com", []string{"parfüm"}, true)

	code := doAuthed(t, r, sess, http.MethodDelete, "/auth/account", nil, nil)
	mustStatus(t, code, statusOK, "hesap silme")

	for _, table := range []string{"user_profiles", "user_allergens", "consent_log"} {
		var n int
		q := "SELECT count(*) FROM " + table + " WHERE "
		if table == "user_profiles" {
			q += "id = $1"
		} else {
			q += "user_id = $1"
		}
		if err := db.QueryRow(q, userID).Scan(&n); err != nil {
			t.Fatalf("%s sayımı başarısız: %v", table, err)
		}
		if n != 0 {
			t.Errorf("%s tablosunda %d satır kaldı", table, n)
		}
	}
}

// TestLoginDoesNotLeakAccountExistence, olmayan hesap ile yanlış parolanın
// aynı yanıtı verdiğini doğrular. Aksi halde uç nokta, hangi e-postaların
// kayıtlı olduğunu sızdıran bir araca dönüşür.
func TestLoginDoesNotLeakAccountExistence(t *testing.T) {
	_, r, _ := newTestServer(t)

	registerUser(t, r, "var@ornek.com", nil, false)

	var wrongPassword, noAccount struct {
		Error string `json:"error"`
	}

	code1 := doJSON(t, r, http.MethodPost, "/auth/login", map[string]any{
		"email": "var@ornek.com", "password": "yanlis-parola",
	}, &wrongPassword)
	code2 := doJSON(t, r, http.MethodPost, "/auth/login", map[string]any{
		"email": "yok@ornek.com", "password": "yanlis-parola",
	}, &noAccount)

	if code1 != code2 {
		t.Errorf("durum kodları farklı: yanlış parola %d, olmayan hesap %d", code1, code2)
	}
	if wrongPassword.Error != noAccount.Error {
		t.Errorf("hata mesajları hesabın varlığını sızdırıyor: %q vs %q",
			wrongPassword.Error, noAccount.Error)
	}
}

// TestLoginSucceeds, doğru bilgilerle girişin oturum açtığını doğrular.
func TestLoginSucceeds(t *testing.T) {
	_, r, _ := newTestServer(t)

	registerUser(t, r, "giris@ornek.com", nil, false)

	raw, _ := json.Marshal(map[string]any{
		"email": "giris@ornek.com", "password": "cok-guclu-bir-parola",
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(string(raw)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	mustStatus(t, w.Code, statusOK, "giriş")

	var found bool
	for _, ck := range w.Result().Cookies() {
		if ck.Name == "clarity_session" {
			found = true
			if !ck.HttpOnly {
				t.Error("oturum cookie'si httpOnly değil — XSS ile çalınabilir")
			}
		}
	}
	if !found {
		t.Error("giriş sonrası oturum cookie'si gelmedi")
	}
}

// TestWeakPasswordRejected, kısa parolanın reddedildiğini doğrular.
func TestWeakPasswordRejected(t *testing.T) {
	_, r, _ := newTestServer(t)

	code := doJSON(t, r, http.MethodPost, "/auth/register", map[string]any{
		"email": "zayif@ornek.com", "password": "kisa",
	}, nil)
	mustStatus(t, code, http.StatusBadRequest, "zayıf parola")
}

// TestDuplicateEmailRejected, aynı e-postayla ikinci hesabın açılamadığını
// doğrular.
func TestDuplicateEmailRejected(t *testing.T) {
	_, r, _ := newTestServer(t)

	registerUser(t, r, "tekrar@ornek.com", nil, false)

	code := doJSON(t, r, http.MethodPost, "/auth/register", map[string]any{
		"email": "tekrar@ornek.com", "password": "cok-guclu-bir-parola",
	}, nil)
	mustStatus(t, code, http.StatusConflict, "tekrar eden e-posta")
}
