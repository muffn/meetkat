package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfCookieName = "meetkat_csrf"
const csrfContextKey = "csrf_token"

// CSRF implements the Double Submit Cookie pattern.
// A random token is stored in an HttpOnly cookie and injected into the Gin context.
// State-changing requests must present the same token via X-CSRF-Token header or
// a csrf_token form field.
func CSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(csrfCookieName)
		if err != nil || token == "" {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			token = hex.EncodeToString(b)
			secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
			c.SetCookie(csrfCookieName, token, 0, "/", "", secure, true)
		}
		c.Set(csrfContextKey, token)

		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			submitted := c.GetHeader("X-CSRF-Token")
			if submitted == "" {
				submitted = c.PostForm("csrf_token")
			}
			if submitted == "" || submitted != token {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
		}
		c.Next()
	}
}
