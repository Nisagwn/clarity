// middleware paketi, enine kesen HTTP ara katmanlarını barındırır.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS, Next.js geliştirme sunucusunun (ve yapılandırılmış diğer kaynakların)
// API'yi tarayıcıdan çağırmasına izin verir. Virgülle ayrılmış kaynak listesi
// alır; "*" tümüne izin verir.
func CORS(allowedOrigins string) gin.HandlerFunc {
	allowAll := strings.TrimSpace(allowedOrigins) == "*" || strings.TrimSpace(allowedOrigins) == ""

	allowed := map[string]bool{}
	for _, o := range strings.Split(allowedOrigins, ",") {
		if o = strings.TrimSpace(o); o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Oturum httpOnly cookie ile taşınıyor. Kimlik bilgisi taşıyan
		// isteklerde "*" geçersizdir: tarayıcı, kaynağın tam olarak
		// belirtilmesini şart koşar. Bu yüzden allowAll durumunda bile
		// isteğin kendi kaynağını yansıtıyoruz.
		switch {
		case origin != "" && (allowAll || allowed[origin]):
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		case allowAll:
			// Kaynak başlığı yok (ör. curl): kimlik bilgisi de gerekmiyor.
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
