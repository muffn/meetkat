package handler

import (
	"meetkat/internal/i18n"

	"github.com/gin-gonic/gin"
)

// LocalizerFromCtx retrieves the *i18n.Localizer set by the language middleware.
func LocalizerFromCtx(c *gin.Context) *i18n.Localizer {
	return c.MustGet("localizer").(*i18n.Localizer)
}
