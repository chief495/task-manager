package user

import (
	"log"
	"net/http"
	"strconv"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/gin-gonic/gin"
)

// UserHandler — обработчик HTTP‑запросов для пользователей
type UserHandler struct {
	service *UserService
}

// NewUserHandler — конструктор обработчика
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

// CreateUser — создание нового пользователя
// @Summary Создать пользователя
// @Description Создаёт нового пользователя с указанными данными
// @Tags пользователи
// @Accept json
// @Produce json
// @Param user body struct{Email string `json:"email" binding:"required,email"` Password string `json:"password" binding:"required,min=6"` Name string `json:"name"`} true "Данные пользователя"
// @Success 201 {object} models.User "Созданный пользователь"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var input CreateUserRequest
	// Валидация входных данных
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создание модели пользователя
	user := &models.User{
		Email:    input.Email,
		Password: input.Password,
		Name:     input.Name,
		Role:     input.Role,
	}

	// Вызов сервиса
	err := h.service.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, NewUserResponse(user))
}

// GetUserTasksStats — статистика по задачам пользователя
// @Summary Получить статистику по задачам пользователя
// @Description Возвращает количество задач пользователя по статусам (всего, в работе, завершено)
// @Tags пользователи
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} struct{UserID uint `json:"user_id"` Total int `json:"total"` InProgress int `json:"in_progress"` Done int `json:"done"`} "Статистика по задачам"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/{id}/task-stats [get]
func (h *UserHandler) GetUserTasksStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	stats, err := h.service.GetUserTaskStats(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     stats.UserID,
		"total":       stats.Total,
		"in_progress": stats.InProgress,
		"done":        stats.Done,
	})
}

// GetUser — получение пользователя по ID
// @Summary Получить пользователя по ID
// @Description Возвращает информацию о пользователе с указанным ID
// @Tags пользователи
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.User "Пользователь"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 404 {object} gin.H "Пользователь не найден"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	// Получаем ID из URL
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	// Преобразуем строку в uint64
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	// Вызов сервиса
	user, err := h.service.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, NewUserResponse(user))
}

// ListUsers — получение списка всех пользователей
// @Summary Получить список всех пользователей
// @Description Возвращает список всех зарегистрированных пользователей
// @Tags пользователи
// @Produce json
// @Success 200 {array} models.User "Список пользователей"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, NewUserListResponse(users))
}

// GetUserSocialRatingHistory — получение истории социального рейтинга пользователя
// @Summary Получить историю социального рейтинга
// @Description Возвращает историю изменения социального рейтинга пользователя за указанное количество дней
// @Tags пользователи
// @Produce json
// @Param id path int true "ID пользователя"
// @Param days query int false "Количество дней (по умолчанию 14)" default(14)
// @Success 200 {array} models.SocialRatingLog "История рейтинга"
// @Failure 400 {object} gin.H "Некорректный формат ID или параметра days"
// @Failure 404 {object} gin.H "Пользователь не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/{id}/social-rating-history [get]
func (h *UserHandler) GetUserSocialRatingHistory(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	daysStr := c.DefaultQuery("days", "14")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid days parameter"})
		return
	}

	if days > 365 {
		days = 365 // Ограничиваем максимальный период
	}

	history, err := h.service.GetUserSocialRatingHistory(uint(id), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, history)
}

// UpdateUser — обновление пользователя
// @Summary Обновить пользователя
// @Description Обновляет данные пользователя по ID
// @Tags пользователи
// @Accept json
// @Produce json
// @Param user body struct{ID uint `json:"id"` Email string `json:"email" binding:"required,email"` Password string `json:"password"` Name string `json:"name"`} true "Данные для обновления"
// @Success 2 gef 200 {object} models.User "Обновлённый пользователь"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 404 {object} gin.H "Пользователь не найден"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	var input UpdateUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Println(input)
	// Получаем ID из URL
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	// Преобразуем строку в uint64
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// Вызов сервиса
	user, err := h.service.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	// Обновление полей
	user.Email = input.Email
	user.Name = input.Name
	if input.Password != "" {
		user.Password = input.Password // Сервис сам хеширует пароль
	}
	// Парсинг float32 из строки
	middlePlanTime, err := strconv.ParseFloat(input.MiddlePlanTimeStr, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid middle_plan_time: must be a valid number")
		return
	}
	user.MiddlePlanTime = float32(middlePlanTime)
	socialRating, err := strconv.ParseFloat(input.SocialRatingStr, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, "invalid social_rating: must be a valid number")
		return
	}
	user.SocialRating = float32(socialRating)
	user.Role = input.Role
	log.Println(user)
	err = h.service.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser — удаление пользователя
// @Summary Удалить пользователя
// @Description Удаляет пользователя по ID
// @Tags пользователи
// @Param id path int true "ID пользователя"
// @Success 204 "Пользователь удалён"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	err = h.service.DeleteUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// AuthenticateUser — аутентификация пользователя
// @Summary Аутентифицировать пользователя
// @Description Выполняет вход пользователя в систему
// @Tags пользователи
// @Accept json
// @Produce json
// @Param credentials body struct{Email string `json:"email" binding:"required,email"` Password string `json:"password" binding:"required"`} true "Данные для аутентификации"
// @Success 200 {object} models.User "Данные аутентифицированного пользователя"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 401 {object} gin.H "Неверные email или пароль"
// @Router /users/authenticate [post]
func (h *UserHandler) AuthenticateUser(c *gin.Context) {
	var input AuthenticateUserRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.AuthenticateUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword — смена пароля
// @Summary Сменить пароль пользователя
// @Description Позволяет пользователю сменить пароль
// @Tags пользователи
// @Accept json
// @Produce json
// @Param password_change body struct{UserID uint `json:"user_id"` OldPassword string `json:"old_password" binding:"required"` NewPassword string `json:"new_password" binding:"required,min=6"`} true "Данные для смены пароля"
// @Success 200 {object} gin.H "Сообщение об успешной смене пароля"
// @Failure 400 {object} gin.H "Ошибка валидации входных данных"
// @Failure 401 {object} gin.H "Неверный старый пароль"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/change-password [post]
func (h *UserHandler) ChangePassword(c *gin.Context) {
	var input ChangePasswordRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.ChangePassword(input.UserID, input.OldPassword, input.NewPassword)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}

// ActivateUser — активация пользователя
// @Summary Активировать пользователя
// @Description Активирует пользователя по ID (например, после регистрации)
// @Tags пользователи
// @Param id path int true "ID пользователя"
// @Success 200 {object} gin.H "Сообщение об активации"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/{id}/activate [post]
func (h *UserHandler) ActivateUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
		return
	}

	err = h.service.ActivateUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user activated"})
}

// DeactivateUser — деактивация пользователя
// @Summary Деактивировать пользователя
// @Description Деактивирует пользователя по ID (блокировка аккаунта)
// @Tags пользователи
// @Param id path int true "ID пользователя"
// @Success 200 {object} gin.H "Сообщение о деактивации"
// @Failure 400 {object} gin.H "Некорректный формат ID"
// @Failure 500 {object} gin.H "Внутренняя ошибка сервера"
// @Router /users/{id}/deactivate [post]
func (h *UserHandler) DeactivateUser(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	err = h.service.DeactivateUser(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deactivated"})
}
