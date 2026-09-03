package middleware

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Oturum, httpOnly bir cookie'de taşınan JWT ile tutulur.
//
// Neden cookie, neden Authorization başlığı değil: jetonu JavaScript'ten
// okunamayacak yere koymak, XSS durumunda oturumun çalınmasını zorlaştırır.
// Sağlık verisi taşıyan bir uygulamada bu fark önemli.

const (
	// SessionCookie, oturum jetonunun taşındığı cookie adı.
	SessionCookie = "clarity_session"

	// sessionTTL, jetonun geçerlilik süresi.
	sessionTTL = 7 * 24 * time.Hour

	// contextUserID, doğrulanmış kullanıcı kimliğinin gin context'indeki anahtarı.
	contextUserID = "user_id"
)

// jwtSecret, imzalama anahtarını döndürür. Tanımsızsa boş döner ve
// çağıran taraf hata verir — üretimde sessizce zayıf bir anahtara düşmek
// oturum güvenliğini sessizce kaldırmak olurdu.
func jwtSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// IssueSession, kullanıcı için bir oturum jetonu üretir ve cookie'ye yazar.
func IssueSession(c *gin.Context, userID int, secureCookie bool) error {
	secret := jwtSecret()
	if len(secret) == 0 {
		return errMissingSecret
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(sessionTTL)),
	})

	signed, err := token.SignedString(secret)
	if err != nil {
		return err
	}

	// SameSite=Lax: form gönderimlerinde çalışır ama siteler arası
	// isteklerde jeton gitmez (CSRF yüzeyi daralır).
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, signed, int(sessionTTL.Seconds()), "/", "", secureCookie, true)
	return nil
}

// ClearSession, oturum cookie'sini siler.
func ClearSession(c *gin.Context, secureCookie bool) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(SessionCookie, "", -1, "/", "", secureCookie, true)
}

// RequireAuth, geçerli bir oturum yoksa isteği 401 ile keser.
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := parseSession(c)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Bu işlem için giriş yapmalısınız",
			})
			return
		}
		c.Set(contextUserID, userID)
		c.Next()
	}
}

// OptionalAuth, oturum varsa kullanıcıyı context'e koyar, yoksa devam eder.
// Giriş yapmadan da kullanılabilen uç noktalar için.
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if userID, ok := parseSession(c); ok {
			c.Set(contextUserID, userID)
		}
		c.Next()
	}
}

// UserID, doğrulanmış kullanıcı kimliğini döndürür.
func UserID(c *gin.Context) (int, bool) {
	v, exists := c.Get(contextUserID)
	if !exists {
		return 0, false
	}
	id, ok := v.(int)
	return id, ok
}

// parseSession, cookie'deki jetonu doğrular.
func parseSession(c *gin.Context) (int, bool) {
	secret := jwtSecret()
	if len(secret) == 0 {
		return 0, false
	}

	raw, err := c.Cookie(SessionCookie)
	if err != nil || raw == "" {
		return 0, false
	}

	var claims jwt.RegisteredClaims
	// İmza yöntemi açıkça kısıtlanır: aksi halde "alg: none" saldırısına
	// açık kalırdık.
	_, err = jwt.ParseWithClaims(raw, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errUnexpectedSigningMethod
		}
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, false
	}

	userID, err := strconv.Atoi(claims.Subject)
	if err != nil || userID <= 0 {
		return 0, false
	}
	return userID, true
}

type authError string

func (e authError) Error() string { return string(e) }

const (
	errMissingSecret           = authError("JWT_SECRET tanımlı değil")
	errUnexpectedSigningMethod = authError("beklenmeyen imza yöntemi")
)
