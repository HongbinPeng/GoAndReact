package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"photo-album-server/internal/util"
)

const currentUserKey = "currentUser"

type AuthUser struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			util.Error(c, 401, "没有提供Authorization请求头")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			util.Error(c, 401, "Authorization请求头格式错误，应为 Bearer {token}")
			return
		}

		claims, err := util.ParseToken(parts[1], secret)
		if err != nil {
			util.Error(c, 401, "登录凭证无效或已过期")
			return
		}

		c.Set(currentUserKey, AuthUser{
			ID:       claims.UserID,
			Username: claims.Username,
		})
		c.Next()
	}
}

func GetCurrentUser(c *gin.Context) (AuthUser, bool) {
	value, exists := c.Get(currentUserKey)
	if !exists {
		return AuthUser{}, false
	}

	user, ok := value.(AuthUser)
	return user, ok
}
