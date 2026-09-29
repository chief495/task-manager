package user

import "github.com/eduardnesterov/task-manager/internal/models"

// CreateUserRequest — входящий запрос на создание пользователя
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// RegisterRequest - запрос на регистрацию
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// UpdateUserRequest — входящий запрос на обновление
type UpdateUserRequest struct {
	ID                uint   `json:"id"`
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password"`
	Name              string `json:"name"`
	Role              string `json:"role"`
	IsActive          *bool  `json:"is_active"`
	MiddlePlanTimeStr string `json:"middle_plan_time"`
	SocialRatingStr   string `json:"social_rating"`
}

// LoginRequest — запрос на логин
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse — ответ клиенту (без пароля)
type UserResponse struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	Email          string  `json:"email"`
	Role           string  `json:"role"`
	IsActive       bool    `json:"is_active"`
	SocialRating   float32 `json:"social_rating"`
	MiddlePlanTime float32 `json:"middle_plan_time"`
}

func NewUserResponse(u *models.User) UserResponse { // ← models.User
	return UserResponse{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		Role:           u.Role,
		IsActive:       u.IsActive,
		SocialRating:   u.SocialRating,
		MiddlePlanTime: u.MiddlePlanTime,
	}
}

func NewUserListResponse(users []*models.User) []UserResponse { // ← *models.User
	result := make([]UserResponse, len(users))
	for i := range users {
		result[i] = NewUserResponse(users[i])
	}
	return result
}

type ResetPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type AuthenticateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	UserID      uint   `json:"user_id"`
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}
