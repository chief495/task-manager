package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware — проверяет JWT-токен в заголовке Authorization или в cookie "token"
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""

		// 1. Пробуем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// 2. Если заголовка нет — пробуем cookie
		if tokenString == "" {
			if cookie, err := c.Cookie("token"); err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		// 3. Токен не найден ни там ни там
		if tokenString == "" {
			if isHTMLRequest(c) {
				c.Redirect(http.StatusTemporaryRedirect, "/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			}
			c.Abort()
			return
		}

		// 4. Парсим и валидируем токен
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			if isHTMLRequest(c) {
				// Сбрасываем невалидный токен и редиректим на логин
				secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
				c.SetCookie("token", "", -1, "/", "", secure, true)
				c.Redirect(http.StatusTemporaryRedirect, "/login")
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			}
			c.Abort()
			return
		}

		// 5. Прокидываем данные пользователя в контекст
		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// isHTMLRequest определяет, ожидает ли клиент HTML.
// Браузер при переходе по URL шлёт Accept: text/html,
// fetch/axios по умолчанию шлют application/json или */*
func isHTMLRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "text/html")
}
