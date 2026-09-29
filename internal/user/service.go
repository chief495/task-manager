package user

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/platform/email"
	"golang.org/x/crypto/bcrypt"
)

// UserService — сервис для работы с пользователями
type UserService struct {
	repo         UserRepository
	emailService *email.EmailService // добавляем email сервис
}

// NewUserService — конструктор сервиса
func NewUserService(repo UserRepository, emailService *email.EmailService) *UserService {
	return &UserService{
		repo:         repo,
		emailService: emailService,
	}
}

// CreateUser — создание нового пользователя
// Пароль автоматически хешируется
func (s *UserService) CreateUser(user *models.User) error {
	if user.Email == "" {
		return errors.New("email не может быть пустым")
	}
	if user.Password == "" {
		return errors.New("пароль не может быть пустым")
	}
	if !isValidEmail(user.Email) {
		return errors.New("некорректный email")
	}

	// Хешируем пароль
	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return err
	}
	user.Password = hashedPassword

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return s.repo.Create(user)
}

// GetUserTaskStats — сбор статистики по задачам пользователя через репозиторий
func (s *UserService) GetUserTaskStats(userID uint) (*models.TaskStats, error) {
	return s.repo.GetUserTaskStats(userID)
}

// GetUserSocialRatingHistory возвращает историю социального рейтинга пользователя за последние N дней
func (s *UserService) GetUserSocialRatingHistory(userID uint, days int) ([]models.SocialRatingLog, error) {
	var history []models.SocialRatingLog

	startDate := time.Now().AddDate(0, 0, -days)

	if err := s.repo.GetDB().
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Order("created_at ASC").
		Find(&history).Error; err != nil {
		return nil, err
	}

	return history, nil
}

// GetUser — получение пользователя по ID
func (s *UserService) GetUser(id uint) (*models.User, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.User), nil
}

// GetUserByEmail — получение пользователя по email
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email не может быть пустым")
	}

	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser — обновление пользователя
// Пароль при обновлении также хешируется
func (s *UserService) UpdateUser(user *models.User) error {
	if user.ID == 0 {
		return errors.New("ID пользователя не указан")
	}
	if user.Email == "" {
		return errors.New("email не может быть пустым")
	}
	if !isValidEmail(user.Email) {
		return errors.New("некорректный email")
	}

	user.UpdatedAt = time.Now()
	return s.repo.Update(user)
}

// DeleteUser — удаление пользователя
func (s *UserService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// ListUsers — получение списка всех пользователей
func (s *UserService) ListUsers() ([]*models.User, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, len(entities))
	for i, e := range entities {
		users[i] = e.(*models.User)
	}
	return users, nil
}

// AuthenticateUser — аутентификация пользователя (email + пароль)
// Возвращает пользователя и ошибку
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	if email == "" || password == "" {
		return nil, errors.New("email и пароль не могут быть пустыми")
	}

	user, err := s.GetUserByEmail(email)
	if err != nil {
		return nil, errors.New("пользователь не найден")
	}

	if !checkPasswordHash(password, user.Password) {
		return nil, errors.New("неверный пароль")
	}

	return user, nil
}

// ChangePassword — смена пароля пользователя
func (s *UserService) ChangePassword(id uint, oldPassword, newPassword string) error {
	if id == 0 {
		return errors.New("ID пользователя не указан")
	}

	user, err := s.GetUser(id)
	if err != nil {
		return err
	}

	// Если пароль не передан — генерируем новый
	password := newPassword
	if password == "" {
		password, err = generatePassword(12)
		if err != nil {
			return fmt.Errorf("ошибка генерации пароля: %w", err)
		}
	}

	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("ошибка хеширования пароля: %w", err)
	}

	// Обновляем пароль напрямую через repo, минуя повторное хеширование в UpdateUser
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()
	if err = s.repo.Update(user); err != nil {
		return fmt.Errorf("ошибка обновления пользователя: %w", err)
	}

	// Отправляем новый пароль на почту
	if err = s.emailService.SendNewPassword(user.Email, password); err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}

	return nil
}

// ResetPassword — сброс пароля (для админов или восстановления)
func (s *UserService) ResetPassword(id uint, newPassword string) error {
	if id == 0 {
		return errors.New("ID пользователя не указан")
	}
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}
	if newPassword == "" {
		newPassword, err = generatePassword(8)
	}
	if err != nil {
		return err
	}
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	// Отправляем новый пароль на почту
	if err = s.emailService.SendNewPassword(user.Email, newPassword); err != nil {
		return fmt.Errorf("ошибка отправки email: %w", err)
	}

	user.Password = hashedPassword
	user.UpdatedAt = time.Now()
	return s.UpdateUser(user)
}

// ActivateUser — активация пользователя
func (s *UserService) ActivateUser(id uint) error {
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}

	user.IsActive = true
	user.UpdatedAt = time.Now()
	return s.UpdateUser(user)
}

// DeactivateUser — деактивация пользователя
func (s *UserService) DeactivateUser(id uint) error {
	user, err := s.GetUser(id)
	if err != nil {
		return err
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()
	return nil
}

// SearchUsers — поиск пользователей по email или имени
func (s *UserService) SearchUsers(query string) ([]*models.User, error) {
	if query == "" {
		return nil, errors.New("запрос не может быть пустым")
	}

	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}

	users := make([]*models.User, len(entities))
	for i, e := range entities {
		users[i] = e.(*models.User)
	}
	return users, nil
}

// --- Вспомогательные функции ---

// isValidEmail — проверка корректности email
func isValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// hashPassword — хеширование пароля с помощью bcrypt
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// checkPasswordHash — проверка пароля против хеша
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// generatePassword — криптографически стойкая генерация пароля
func generatePassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%"
	password := make([]byte, length)

	for i := range password {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("ошибка генерации символа пароля: %w", err)
		}
		password[i] = charset[idx.Int64()]
	}

	return string(password), nil
}
