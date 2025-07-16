package middleware

import (
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// SessionOpts 设置选项
type SessionOpts struct {
	Name   string
	Secure bool   // required
	Secret []byte // required
	// redis store
	RedisAddr string
	RedisPwd  string
}

// SessionMiddleware session中间件
func SessionMiddleware(opts SessionOpts) gin.HandlerFunc {
	store := cookie.NewStore(opts.Secret)
	store.Options(sessions.Options{
		MaxAge:   86400 * 30,
		Path:     "/",
		Secure:   opts.Secure,
		HttpOnly: true,
	})
	name := "SESSIONID"
	if opts.Name != "" {
		name = opts.Name
	}
	return sessions.Sessions(name, store)
}

// UserMiddleware 用户cookie标记
func UserMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("u")
		if err != nil || cookie == "" {
			u1 := uuid.NewV4().String()
			c.SetCookie("u", u1, 86400*730, "/", "", true, true)
		}
	}
}

// AuthFilter auth filter
func AuthFilter(c *gin.Context) {
	if !IsLogined(c) {
		c.Abort()
		c.Status(http.StatusUnauthorized)
		c.Redirect(http.StatusFound, "/admin/login")
		return
	}

	c.Next()
}

// SetLogin login user
func SetLogin(c *gin.Context, username string) {
	session := sessions.Default(c)
	session.Set("username", username)
	session.Save()
}

// SetLogout logout user
func SetLogout(c *gin.Context) {
	session := sessions.Default(c)
	session.Delete("username")
	session.Save()
}

// IsLogined account logined
func IsLogined(c *gin.Context) bool {
	return GetUsername(c) != ""
}

// GetUsername get logined account
func GetUsername(c *gin.Context) string {
	session := sessions.Default(c)
	username := session.Get("username")
	if username == nil {
		return ""
	}
	return username.(string)
}
