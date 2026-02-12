package middleware

import (
	"meetkat/internal/i18n"

	"github.com/gin-gonic/gin"
)

// LangCookie returns middleware that resolves the user's language preference.
// Priority: ?lang= query param > meetkat_lang cookie > Accept-Language header.
// When ?lang= is present the cookie is set/updated.
func LangCookie(tr *i18n.Translator) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.Query("lang")
		if lang != "" {
			c.SetCookie("meetkat_lang", lang, 365*24*60*60, "/", "", false, false)
		} else {
			lang, _ = c.Cookie("meetkat_lang")
			if lang == "" {
				lang = tr.Match(c.GetHeader("Accept-Language"))
			}
		}
		loc := tr.ForLang(lang)
		c.Set("localizer", loc)
		c.Next()
	}
}
