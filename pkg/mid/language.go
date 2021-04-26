// Package mid provides ...
package mid

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// LangOpts 语言选项
type LangOpts struct {
	CookieName string
	Default    string
	Supported  []string
}

// isExist language
func (opts LangOpts) isExist(l string) bool {
	for _, v := range opts.Supported {
		if v == l {
			return true
		}
	}
	return false
}

// LangMiddleware set language
func LangMiddleware(opts LangOpts) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang, err := c.Cookie(opts.CookieName)
		// found cookie
		if err == nil {
			c.Set(opts.CookieName, lang)
			return
		}
		// set cookie
		al := strings.ToLower(c.GetHeader("Accept-Language"))
		if al != "" {
			// choose default if not supported
			lang = opts.Default

			langs := strings.Split(al, ",")
			for _, v := range langs {
				if opts.isExist(v) {
					lang = v
					break
				}
			}
		} else {
			lang = opts.Default
		}
		c.SetCookie(opts.CookieName, lang, 86400*365, "/", "", false, false)
		c.Set(opts.CookieName, lang)
	}
}
