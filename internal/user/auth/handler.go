package auth

import (
	"log"
	"net/http"

	"github.com/eduardnesterov/task-manager/internal/models"
	platformauth "github.com/eduardnesterov/task-manager/internal/platform/auth"
	"github.com/eduardnesterov/task-manager/internal/user"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userService *user.UserService
}

func NewAuthHandler(userService *user.UserService) *AuthHandler {
	return &AuthHandler{userService: userService}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var input user.LoginRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	user, err := h.userService.GetUserByEmail(input.Email)
	if err != nil || user == nil || !user.CheckPassword(input.Password) {
		// Одинаковое сообщение для всех случаев ошибки
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := platformauth.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetCookie("token", token, 3600, "/", "", secure, true)
	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": user.ID,
		"role":    user.Role,
	})
}

// Register — регистрация нового пользователя (для гостей)
func (h *AuthHandler) Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     "user",
		IsActive: true,
	}

	if err := h.userService.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created"})
}

// ResetPassword — сброс пароля по email пользователя
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input user.ResetPasswordRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Ищем пользователя по email
	user, err := h.userService.GetUserByEmail(input.Email)
	if err != nil || user == nil {
		// Не раскрываем существование email
		c.JSON(http.StatusOK, gin.H{
			"message": "Если email существует, новый пароль будет отправлен на почту",
		})
		return
	}

	// Передаём пустой пароль — сервис сам сгенерирует новый
	if err = h.userService.ResetPassword(user.ID, ""); err != nil {
		log.Printf("reset password error for user %d: %v", user.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сбросить пароль"})
		return
	}

	// Одинаковый ответ для защиты от перебора email
	c.JSON(http.StatusOK, gin.H{
		"message": "Если email существует, новый пароль будет отправлен на почту",
	})
}
