package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	platformauth "github.com/eduardnesterov/task-manager/internal/platform/auth"
	"github.com/eduardnesterov/task-manager/internal/project"
	"github.com/eduardnesterov/task-manager/internal/project/participant"
	"github.com/eduardnesterov/task-manager/internal/project/system"
	"github.com/eduardnesterov/task-manager/internal/project/trip"
	"github.com/eduardnesterov/task-manager/internal/report"
	"github.com/eduardnesterov/task-manager/internal/task"
	"github.com/eduardnesterov/task-manager/internal/user"
	userauth "github.com/eduardnesterov/task-manager/internal/user/auth"
	"github.com/eduardnesterov/task-manager/internal/workflow/status"
	status_transition "github.com/eduardnesterov/task-manager/internal/workflow/status/transition"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthMiddlewareRejectsMissingAuthorization(t *testing.T) {
	router := gin.New()
	router.Use(platformauth.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestAuthMiddlewareAllowsValidBearerToken(t *testing.T) {
	token, err := platformauth.GenerateJWT(42, "user@example.com", "manager")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	router := gin.New()
	router.Use(platformauth.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"user_id": c.MustGet("user_id"),
			"email":   c.MustGet("email"),
			"role":    c.MustGet("role"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"user_id":42`) || !strings.Contains(body, `"role":"manager"`) {
		t.Fatalf("unexpected response body: %s", body)
	}
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userRepo := newFakeUserRepo(&models.User{
		Model:    gorm.Model{ID: 1},
		ID:       1,
		Email:    "admin@example.com",
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	})
	authHandler := userauth.NewAuthHandler(user.NewUserService(userRepo, nil))

	router := gin.New()
	router.POST("/login", authHandler.Login)

	body := `{"email":"admin@example.com","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	responseBody := rec.Body.String()
	if !strings.Contains(responseBody, `"token":"`) {
		t.Fatalf("expected JWT token in response, got: %s", responseBody)
	}
	if !strings.Contains(responseBody, `"user_id":1`) {
		t.Fatalf("expected user_id in response, got: %s", responseBody)
	}
}

func TestUserHandlerGetUserDoesNotExposePassword(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userRepo := newFakeUserRepo(&models.User{
		Model:    gorm.Model{ID: 1},
		ID:       1,
		Email:    "admin@example.com",
		Name:     "Admin",
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	})
	userHandler := user.NewUserHandler(user.NewUserService(userRepo, nil))

	router := gin.New()
	router.GET("/users/:id", userHandler.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"password"`) {
		t.Fatalf("response must not contain password field: %s", rec.Body.String())
	}
}

func TestUserHandlerListUsersDoesNotExposePassword(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userRepo := newFakeUserRepo(
		&models.User{
			Model:    gorm.Model{ID: 1},
			ID:       1,
			Email:    "admin@example.com",
			Name:     "Admin",
			Password: string(hashedPassword),
			Role:     "admin",
			IsActive: true,
		},
		&models.User{
			Model:    gorm.Model{ID: 2},
			ID:       2,
			Email:    "user@example.com",
			Name:     "User",
			Password: string(hashedPassword),
			Role:     "user",
			IsActive: true,
		},
	)
	userHandler := user.NewUserHandler(user.NewUserService(userRepo, nil))

	router := gin.New()
	router.GET("/users", userHandler.ListUsers)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), `"password"`) {
		t.Fatalf("response must not contain password field: %s", rec.Body.String())
	}
}

func TestAuthHandlerLoginInvalidPasswordDoesNotLeakSecrets(t *testing.T) {
	const plainPassword = "secret123"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(plainPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	userRepo := newFakeUserRepo(&models.User{
		Model:    gorm.Model{ID: 1},
		ID:       1,
		Email:    "admin@example.com",
		Password: string(hashedPassword),
		Role:     "admin",
		IsActive: true,
	})
	authHandler := userauth.NewAuthHandler(user.NewUserService(userRepo, nil))

	router := gin.New()
	router.POST("/login", authHandler.Login)

	body := `{"email":"admin@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, rec.Code, rec.Body.String())
	}

	responseBody := rec.Body.String()
	if strings.Contains(responseBody, "wrong-password") {
		t.Fatalf("response leaked plain password: %s", responseBody)
	}
	if strings.Contains(responseBody, string(hashedPassword)) {
		t.Fatalf("response leaked stored hash: %s", responseBody)
	}
}

func TestTaskHandlerCreateTaskWithoutTimesReturnsBadRequest(t *testing.T) {
	taskRepo := newFakeTaskRepo()
	statusRepo := newFakeStatusRepo(&models.Status{Model: gorm.Model{ID: 1}, ID: 1, Name: "В работе"})
	transitionRepo := newFakeStatusTransitionRepo()
	userRepo := newFakeUserRepo(&models.User{Model: gorm.Model{ID: 99}, ID: 99, Email: "executor@example.com"})

	taskHandler := task.NewTaskHandler(task.NewTaskService(taskRepo, statusRepo, transitionRepo, userRepo, nil))

	router := gin.Default()
	router.POST("/tasks", taskHandler.CreateTask)

	reqBody := `{"name":"Review API","status_id":1,"executor_id":99,"project_id":100}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestTaskHandlerUpdateTaskStatusSuccess(t *testing.T) {
	taskRepo := newFakeTaskRepo(&models.Task{
		Model:      gorm.Model{ID: 7},
		ID:         7,
		Name:       "Review API",
		StatusID:   1,
		ExecutorID: 99,
		ProjectID:  100,
	})
	statusRepo := newFakeStatusRepo(
		&models.Status{Model: gorm.Model{ID: 1}, ID: 1, Name: "В работе"},
		&models.Status{Model: gorm.Model{ID: 2}, ID: 2, Name: "Готово", IsFinal: true},
	)
	transitionRepo := newFakeStatusTransitionRepo()
	transitionRepo.allow(1, 2)
	userRepo := newFakeUserRepo(&models.User{Model: gorm.Model{ID: 99}, ID: 99, Email: "executor@example.com"})

	taskHandler := task.NewTaskHandler(task.NewTaskService(taskRepo, statusRepo, transitionRepo, userRepo, nil))

	router := gin.New()
	router.PUT("/tasks/:id/status", taskHandler.UpdateTaskStatus)

	req := httptest.NewRequest(http.MethodPut, "/tasks/7/status", strings.NewReader(`{"status_id":2}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	updated := taskRepo.tasks[7]
	if updated.StatusID != 2 {
		t.Fatalf("expected status_id to be updated to 2, got %d", updated.StatusID)
	}
	if !updated.Completed {
		t.Fatalf("expected task to be marked completed")
	}
}

func TestTaskHandlerUpdateTaskStatusInvalidTransition(t *testing.T) {
	taskRepo := newFakeTaskRepo(&models.Task{
		Model:      gorm.Model{ID: 8},
		ID:         8,
		Name:       "Review API",
		StatusID:   1,
		ExecutorID: 99,
		ProjectID:  100,
	})
	statusRepo := newFakeStatusRepo(
		&models.Status{Model: gorm.Model{ID: 1}, ID: 1, Name: "В работе"},
		&models.Status{Model: gorm.Model{ID: 3}, ID: 3, Name: "Отменено"},
	)
	transitionRepo := newFakeStatusTransitionRepo()
	userRepo := newFakeUserRepo(&models.User{Model: gorm.Model{ID: 99}, ID: 99, Email: "executor@example.com"})

	taskHandler := task.NewTaskHandler(task.NewTaskService(taskRepo, statusRepo, transitionRepo, userRepo, nil))

	router := gin.New()
	router.PUT("/tasks/:id/status", taskHandler.UpdateTaskStatus)

	req := httptest.NewRequest(http.MethodPut, "/tasks/8/status", strings.NewReader(`{"status_id":3}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d: %s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestTaskHandlerUpdateTaskUsesPathID(t *testing.T) {
	taskRepo := newFakeTaskRepo(
		&models.Task{
			Model:      gorm.Model{ID: 1},
			ID:         1,
			Name:       "Task One",
			StatusID:   1,
			ExecutorID: 10,
			ProjectID:  100,
		},
		&models.Task{
			Model:      gorm.Model{ID: 2},
			ID:         2,
			Name:       "Task Two",
			StatusID:   1,
			ExecutorID: 20,
			ProjectID:  200,
		},
	)
	statusRepo := newFakeStatusRepo(&models.Status{Model: gorm.Model{ID: 1}, ID: 1, Name: "В работе"})
	transitionRepo := newFakeStatusTransitionRepo()
	userRepo := newFakeUserRepo(
		&models.User{Model: gorm.Model{ID: 10}, ID: 10, Email: "one@example.com"},
		&models.User{Model: gorm.Model{ID: 20}, ID: 20, Email: "two@example.com"},
	)

	taskHandler := task.NewTaskHandler(task.NewTaskService(taskRepo, statusRepo, transitionRepo, userRepo, nil))

	router := gin.New()
	router.PUT("/tasks/:id", taskHandler.UpdateTask)

	reqBody := `{"id":2,"name":"Renamed via wrong body id"}`
	req := httptest.NewRequest(http.MethodPut, "/tasks/1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if taskRepo.tasks[1].Name != "Renamed via wrong body id" {
		t.Fatalf("expected task from path to be updated, got task[1]=%q", taskRepo.tasks[1].Name)
	}
	if taskRepo.tasks[2].Name != "Task Two" {
		t.Fatalf("expected body id to be ignored, got task[2]=%q", taskRepo.tasks[2].Name)
	}
}

func TestProjectHandlerUpdateProjectUsesPathID(t *testing.T) {
	projectRepo := newFakeProjectRepo(
		&models.Project{
			Model:         gorm.Model{ID: 1},
			ID:            1,
			Name:          "Project One",
			ResponsibleID: 11,
			StageID:       1,
		},
		&models.Project{
			Model:         gorm.Model{ID: 2},
			ID:            2,
			Name:          "Project Two",
			ResponsibleID: 22,
			StageID:       1,
		},
	)
	userRepo := newFakeUserRepo(
		&models.User{Model: gorm.Model{ID: 11}, ID: 11, Email: "owner1@example.com"},
		&models.User{Model: gorm.Model{ID: 22}, ID: 22, Email: "owner2@example.com"},
	)
	projectService := project.NewProjectService(projectRepo, userRepo, &fakeSystemRepo{}, &fakeParticipantRepo{}, &fakeTripRepo{})
	projectHandler := project.NewProjectHandler(projectService)

	router := gin.New()
	router.PUT("/projects/:id", projectHandler.UpdateProject)

	reqBody := `{"id":2,"name":"Renamed via wrong body id"}`
	req := httptest.NewRequest(http.MethodPut, "/projects/1", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if projectRepo.projects[1].Name != "Renamed via wrong body id" {
		t.Fatalf("expected project from path to be updated, got project[1]=%q", projectRepo.projects[1].Name)
	}
	if projectRepo.projects[2].Name != "Project Two" {
		t.Fatalf("expected body id to be ignored, got project[2]=%q", projectRepo.projects[2].Name)
	}
}

func TestReportServiceListByUserReturnsDistinctPointers(t *testing.T) {
	reportRepo := &fakeReportRepo{
		listByUserResult: []models.Report{
			{Model: gorm.Model{ID: 1}, ID: 1, Text: "First", UserID: 10, CreatedAt: time.Now()},
			{Model: gorm.Model{ID: 2}, ID: 2, Text: "Second", UserID: 10, CreatedAt: time.Now().Add(time.Minute)},
		},
	}

	reportService := report.NewReportService(reportRepo)

	reports, err := reportService.ListByUser(10, "")
	if err != nil {
		t.Fatalf("list reports by user: %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(reports))
	}
	if reports[0] == reports[1] {
		t.Fatalf("expected distinct pointers for reports, got identical addresses")
	}
	if reports[0].ID != 1 || reports[1].ID != 2 {
		t.Fatalf("unexpected report IDs: got %d and %d", reports[0].ID, reports[1].ID)
	}
}

type fakeUserRepo struct {
	users map[uint]*models.User
}

func newFakeUserRepo(users ...*models.User) *fakeUserRepo {
	result := &fakeUserRepo{users: make(map[uint]*models.User, len(users))}
	for _, user := range users {
		clone := *user
		result.users[user.ID] = &clone
	}
	return result
}

func (r *fakeUserRepo) Create(entity any) error {
	user, ok := entity.(*models.User)
	if !ok {
		return errors.New("entity must be *models.User")
	}
	if user.ID == 0 {
		user.ID = uint(len(r.users) + 1)
		user.Model.ID = user.ID
	}
	clone := *user
	r.users[user.ID] = &clone
	return nil
}

func (r *fakeUserRepo) GetByID(id uint) (any, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *user
	return &clone, nil
}

func (r *fakeUserRepo) Update(entity any) error {
	user, ok := entity.(*models.User)
	if !ok {
		return errors.New("entity must be *models.User")
	}
	clone := *user
	r.users[user.ID] = &clone
	return nil
}

func (r *fakeUserRepo) Delete(id uint) error {
	delete(r.users, id)
	return nil
}

func (r *fakeUserRepo) List() ([]any, error) {
	result := make([]any, 0, len(r.users))
	for _, user := range r.users {
		clone := *user
		result = append(result, &clone)
	}
	return result, nil
}

func (r *fakeUserRepo) GetDB() *gorm.DB {
	return nil
}

func (r *fakeUserRepo) Count() (int64, error) {
	return int64(len(r.users)), nil
}

func (r *fakeUserRepo) Search(query string) ([]any, error) {
	var result []any
	for _, user := range r.users {
		if strings.Contains(strings.ToLower(user.Name), strings.ToLower(query)) ||
			strings.Contains(strings.ToLower(user.Email), strings.ToLower(query)) {
			clone := *user
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeUserRepo) GetByEmail(email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			clone := *user
			return &clone, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeUserRepo) Exists(id uint) (bool, error) {
	_, ok := r.users[id]
	return ok, nil
}

func (r *fakeUserRepo) GetActiveUsers() ([]models.User, error) {
	var result []models.User
	for _, user := range r.users {
		if user.IsActive {
			result = append(result, *user)
		}
	}
	return result, nil
}

func (r *fakeUserRepo) GetUsersByRole(role string) ([]models.User, error) {
	var result []models.User
	for _, user := range r.users {
		if user.Role == role {
			result = append(result, *user)
		}
	}
	return result, nil
}

func (r *fakeUserRepo) GetUserTaskStats(userID uint) (*models.TaskStats, error) {
	return &models.TaskStats{UserID: userID}, nil
}

type fakeTaskRepo struct {
	tasks map[uint]*models.Task
}

func newFakeTaskRepo(tasks ...*models.Task) *fakeTaskRepo {
	result := &fakeTaskRepo{tasks: make(map[uint]*models.Task, len(tasks))}
	for _, task := range tasks {
		clone := *task
		result.tasks[task.ID] = &clone
	}
	return result
}

func (r *fakeTaskRepo) Create(entity any) error {
	task, ok := entity.(*models.Task)
	if !ok {
		return errors.New("entity must be *models.Task")
	}
	if task.ID == 0 {
		task.ID = uint(len(r.tasks) + 1)
		task.Model.ID = task.ID
	}
	clone := *task
	r.tasks[task.ID] = &clone
	return nil
}

func (r *fakeTaskRepo) GetByID(id uint) (any, error) {
	task, ok := r.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *task
	return &clone, nil
}

func (r *fakeTaskRepo) Update(entity any) error {
	task, ok := entity.(*models.Task)
	if !ok {
		return errors.New("entity must be *models.Task")
	}
	clone := *task
	r.tasks[task.ID] = &clone
	return nil
}

func (r *fakeTaskRepo) Delete(id uint) error {
	delete(r.tasks, id)
	return nil
}

func (r *fakeTaskRepo) List() ([]any, error) {
	result := make([]any, 0, len(r.tasks))
	for _, task := range r.tasks {
		clone := *task
		result = append(result, &clone)
	}
	return result, nil
}

func (r *fakeTaskRepo) GetDB() *gorm.DB {
	return nil
}

func (r *fakeTaskRepo) Count() (int64, error) {
	return int64(len(r.tasks)), nil
}

func (r *fakeTaskRepo) ListByProject(projectID uint) ([]any, error) {
	var result []any
	for _, task := range r.tasks {
		if task.ProjectID == projectID {
			clone := *task
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeTaskRepo) ListByExecutor(executorID uint) ([]models.Task, error) {
	var result []models.Task
	for _, task := range r.tasks {
		if task.ExecutorID == executorID {
			result = append(result, *task)
		}
	}
	return result, nil
}

func (r *fakeTaskRepo) ListActiveByProject(projectID uint) ([]models.Task, error) {
	var result []models.Task
	for _, task := range r.tasks {
		if task.ProjectID == projectID {
			result = append(result, *task)
		}
	}
	return result, nil
}

func (r *fakeTaskRepo) CountByStatus(statusID uint) (int64, error) {
	var count int64
	for _, task := range r.tasks {
		if task.StatusID == statusID {
			count++
		}
	}
	return count, nil
}

func (r *fakeTaskRepo) ListByCategory(category string) ([]any, error) {
	var result []any
	for _, task := range r.tasks {
		if task.Category == category {
			clone := *task
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeTaskRepo) ListByPhase(phaseID uint) ([]any, error) {
	var result []any
	for _, task := range r.tasks {
		if task.ProjectPhaseID != nil && *task.ProjectPhaseID == phaseID {
			clone := *task
			result = append(result, &clone)
		}
	}
	return result, nil
}

type fakeProjectRepo struct {
	projects map[uint]*models.Project
	members  map[uint]*models.ProjectMember
}

func newFakeProjectRepo(projects ...*models.Project) *fakeProjectRepo {
	result := &fakeProjectRepo{
		projects: make(map[uint]*models.Project, len(projects)),
		members:  make(map[uint]*models.ProjectMember),
	}
	for _, project := range projects {
		clone := *project
		result.projects[project.ID] = &clone
	}
	return result
}

func (r *fakeProjectRepo) Create(entity any) error {
	project, ok := entity.(*models.Project)
	if !ok {
		return errors.New("entity must be *models.Project")
	}
	if project.ID == 0 {
		project.ID = uint(len(r.projects) + 1)
		project.Model.ID = project.ID
	}
	clone := *project
	r.projects[project.ID] = &clone
	return nil
}

func (r *fakeProjectRepo) GetByID(id uint) (any, error) {
	project, ok := r.projects[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *project
	return &clone, nil
}

func (r *fakeProjectRepo) Update(entity any) error {
	project, ok := entity.(*models.Project)
	if !ok {
		return errors.New("entity must be *models.Project")
	}
	clone := *project
	r.projects[project.ID] = &clone
	return nil
}

func (r *fakeProjectRepo) Delete(id uint) error {
	delete(r.projects, id)
	return nil
}

func (r *fakeProjectRepo) List() ([]any, error) {
	result := make([]any, 0, len(r.projects))
	for _, project := range r.projects {
		clone := *project
		result = append(result, &clone)
	}
	return result, nil
}

func (r *fakeProjectRepo) GetDB() *gorm.DB {
	return nil
}

func (r *fakeProjectRepo) Count() (int64, error) {
	return int64(len(r.projects)), nil
}

func (r *fakeProjectRepo) Search(query string) ([]any, error) {
	var result []any
	for _, project := range r.projects {
		if strings.Contains(strings.ToLower(project.Name), strings.ToLower(query)) {
			clone := *project
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeProjectRepo) ListByProject(projectID uint) ([]any, error) {
	project, ok := r.projects[projectID]
	if !ok {
		return []any{}, nil
	}
	clone := *project
	return []any{&clone}, nil
}

func (r *fakeProjectRepo) ListByStage(stageID uint) ([]models.Project, error) {
	var result []models.Project
	for _, project := range r.projects {
		if project.StageID == stageID {
			result = append(result, *project)
		}
	}
	return result, nil
}

func (r *fakeProjectRepo) ListByResponsible(responsibleID uint) ([]models.Project, error) {
	var result []models.Project
	for _, project := range r.projects {
		if project.ResponsibleID == responsibleID {
			result = append(result, *project)
		}
	}
	return result, nil
}

func (r *fakeProjectRepo) ListVisibleForUser(userID uint) ([]models.Project, error) {
	var result []models.Project
	for _, project := range r.projects {
		if !project.IsHidden || project.ResponsibleID == userID {
			result = append(result, *project)
		}
	}
	return result, nil
}

func (r *fakeProjectRepo) GetActiveProjects() ([]models.Project, error) {
	var result []models.Project
	for _, project := range r.projects {
		result = append(result, *project)
	}
	return result, nil
}

func (r *fakeProjectRepo) GetProjectsByStage(stageID uint) ([]models.Project, error) {
	return r.ListByStage(stageID)
}

func (r *fakeProjectRepo) CreateMember(entity any) error {
	member, ok := entity.(*models.ProjectMember)
	if !ok {
		return errors.New("entity must be *models.ProjectMember")
	}
	if member.ID == 0 {
		member.ID = uint(len(r.members) + 1)
		member.Model.ID = member.ID
	}
	clone := *member
	r.members[member.ID] = &clone
	return nil
}

func (r *fakeProjectRepo) GetMemberByID(id uint) (*models.ProjectMember, error) {
	member, ok := r.members[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *member
	return &clone, nil
}

func (r *fakeProjectRepo) UpdateMember(member *models.ProjectMember) error {
	clone := *member
	r.members[member.ID] = &clone
	return nil
}

func (r *fakeProjectRepo) DeleteMember(id uint) error {
	delete(r.members, id)
	return nil
}

func (r *fakeProjectRepo) ListMembersByProject(projectID uint) ([]models.ProjectMember, error) {
	var result []models.ProjectMember
	for _, member := range r.members {
		if member.ProjectID == projectID {
			result = append(result, *member)
		}
	}
	return result, nil
}

type fakeStatusRepo struct {
	statuses map[uint]*models.Status
}

func newFakeStatusRepo(statuses ...*models.Status) *fakeStatusRepo {
	result := &fakeStatusRepo{statuses: make(map[uint]*models.Status, len(statuses))}
	for _, status := range statuses {
		clone := *status
		result.statuses[status.ID] = &clone
	}
	return result
}

func (r *fakeStatusRepo) Create(entity any) error {
	status, ok := entity.(*models.Status)
	if !ok {
		return errors.New("entity must be *models.Status")
	}
	clone := *status
	r.statuses[status.ID] = &clone
	return nil
}

func (r *fakeStatusRepo) GetByID(id uint) (any, error) {
	status, ok := r.statuses[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	clone := *status
	return &clone, nil
}

func (r *fakeStatusRepo) Update(entity any) error {
	status, ok := entity.(*models.Status)
	if !ok {
		return errors.New("entity must be *models.Status")
	}
	clone := *status
	r.statuses[status.ID] = &clone
	return nil
}

func (r *fakeStatusRepo) Delete(id uint) error {
	delete(r.statuses, id)
	return nil
}

func (r *fakeStatusRepo) List() ([]any, error) {
	result := make([]any, 0, len(r.statuses))
	for _, status := range r.statuses {
		clone := *status
		result = append(result, &clone)
	}
	return result, nil
}

func (r *fakeStatusRepo) GetDB() *gorm.DB {
	return nil
}

func (r *fakeStatusRepo) Search(query string) ([]any, error) {
	var result []any
	for _, status := range r.statuses {
		if strings.Contains(strings.ToLower(status.Name), strings.ToLower(query)) {
			clone := *status
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeStatusRepo) GetByName(name string) (any, error) {
	for _, status := range r.statuses {
		if status.Name == name {
			clone := *status
			return &clone, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeStatusRepo) IsFinal(statusID uint) (bool, error) {
	status, ok := r.statuses[statusID]
	if !ok {
		return false, gorm.ErrRecordNotFound
	}
	return status.IsFinal, nil
}

func (r *fakeStatusRepo) GetDefaultStatus() (*models.Status, error) {
	for _, status := range r.statuses {
		if status.IsDefault {
			clone := *status
			return &clone, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

type fakeStatusTransitionRepo struct {
	allowed map[[2]uint]bool
}

func newFakeStatusTransitionRepo() *fakeStatusTransitionRepo {
	return &fakeStatusTransitionRepo{allowed: make(map[[2]uint]bool)}
}

func (r *fakeStatusTransitionRepo) allow(fromID, toID uint) {
	r.allowed[[2]uint{fromID, toID}] = true
}

func (r *fakeStatusTransitionRepo) Create(entity any) error { return nil }
func (r *fakeStatusTransitionRepo) GetByID(id uint) (any, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeStatusTransitionRepo) Update(entity any) error { return nil }
func (r *fakeStatusTransitionRepo) Delete(id uint) error    { return nil }
func (r *fakeStatusTransitionRepo) List() ([]any, error)    { return nil, nil }
func (r *fakeStatusTransitionRepo) GetDB() *gorm.DB         { return nil }

func (r *fakeStatusTransitionRepo) GetByStatuses(fromID, toID uint) (*models.StatusTransition, error) {
	if r.allowed[[2]uint{fromID, toID}] {
		return &models.StatusTransition{FromStatusID: fromID, ToStatusID: toID}, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeStatusTransitionRepo) ListByFromStatus(fromID uint) ([]models.StatusTransition, error) {
	var result []models.StatusTransition
	for key := range r.allowed {
		if key[0] == fromID {
			result = append(result, models.StatusTransition{FromStatusID: key[0], ToStatusID: key[1]})
		}
	}
	return result, nil
}

func (r *fakeStatusTransitionRepo) ListByToStatus(toID uint) ([]models.StatusTransition, error) {
	var result []models.StatusTransition
	for key := range r.allowed {
		if key[1] == toID {
			result = append(result, models.StatusTransition{FromStatusID: key[0], ToStatusID: key[1]})
		}
	}
	return result, nil
}

func (r *fakeStatusTransitionRepo) DeleteByStatuses(fromID, toID uint) error {
	delete(r.allowed, [2]uint{fromID, toID})
	return nil
}

func (r *fakeStatusTransitionRepo) CanTransition(fromID, toID uint, task *models.Task) (bool, string, error) {
	ok := r.allowed[[2]uint{fromID, toID}]
	if !ok {
		return false, "transition not allowed", nil
	}
	return true, "transition allowed", nil
}

func (r *fakeStatusTransitionRepo) ExistsTransition(fromID, toID uint) (bool, error) {
	return r.allowed[[2]uint{fromID, toID}], nil
}

type fakeReportRepo struct {
	reports          []*models.Report
	listByUserResult []models.Report
	filterReportsErr error
	listByUserErr    error
}

func (r *fakeReportRepo) Create(entity any) error { return nil }
func (r *fakeReportRepo) GetByID(id uint) (any, error) {
	for _, report := range r.reports {
		if report.ID == id {
			clone := *report
			return &clone, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}
func (r *fakeReportRepo) Update(entity any) error { return nil }
func (r *fakeReportRepo) Delete(id uint) error    { return nil }
func (r *fakeReportRepo) List() ([]any, error)    { return nil, nil }
func (r *fakeReportRepo) GetDB() *gorm.DB         { return nil }
func (r *fakeReportRepo) ListByProject(projectID uint) ([]any, error) {
	var result []any
	for _, report := range r.reports {
		if report.ProjectID == projectID {
			clone := *report
			result = append(result, &clone)
		}
	}
	return result, nil
}

func (r *fakeReportRepo) ListByUser(userID uint, dateStr string) ([]models.Report, error) {
	if r.listByUserErr != nil {
		return nil, r.listByUserErr
	}
	if r.listByUserResult != nil {
		return r.listByUserResult, nil
	}
	var result []models.Report
	for _, report := range r.reports {
		if report.UserID == userID {
			result = append(result, *report)
		}
	}
	return result, nil
}

func (r *fakeReportRepo) FilterReports(dateFrom, dateTo, projectIDStr string, allowedProjectIDs []uint) ([]*models.Report, error) {
	if r.filterReportsErr != nil {
		return nil, r.filterReportsErr
	}
	return r.reports, nil
}

type fakeSystemRepo struct{}

func (r *fakeSystemRepo) Create(entity any) error                     { return nil }
func (r *fakeSystemRepo) GetByID(id uint) (any, error)                { return nil, gorm.ErrRecordNotFound }
func (r *fakeSystemRepo) Update(entity any) error                     { return nil }
func (r *fakeSystemRepo) Delete(id uint) error                        { return nil }
func (r *fakeSystemRepo) List() ([]any, error)                        { return nil, nil }
func (r *fakeSystemRepo) GetDB() *gorm.DB                             { return nil }
func (r *fakeSystemRepo) ListByProject(projectID uint) ([]any, error) { return nil, nil }

type fakeParticipantRepo struct{}

func (r *fakeParticipantRepo) Create(entity any) error                     { return nil }
func (r *fakeParticipantRepo) GetByID(id uint) (any, error)                { return nil, gorm.ErrRecordNotFound }
func (r *fakeParticipantRepo) Update(entity any) error                     { return nil }
func (r *fakeParticipantRepo) Delete(id uint) error                        { return nil }
func (r *fakeParticipantRepo) List() ([]any, error)                        { return nil, nil }
func (r *fakeParticipantRepo) GetDB() *gorm.DB                             { return nil }
func (r *fakeParticipantRepo) ListByProject(projectID uint) ([]any, error) { return nil, nil }

type fakeTripRepo struct{}

func (r *fakeTripRepo) Create(entity any) error                     { return nil }
func (r *fakeTripRepo) GetByID(id uint) (any, error)                { return nil, gorm.ErrRecordNotFound }
func (r *fakeTripRepo) Update(entity any) error                     { return nil }
func (r *fakeTripRepo) Delete(id uint) error                        { return nil }
func (r *fakeTripRepo) List() ([]any, error)                        { return nil, nil }
func (r *fakeTripRepo) GetDB() *gorm.DB                             { return nil }
func (r *fakeTripRepo) Count() (int64, error)                       { return 0, nil }
func (r *fakeTripRepo) ListByProject(projectID uint) ([]any, error) { return nil, nil }
func (r *fakeTripRepo) ListByPerson(personID uint) ([]models.Trip, error) {
	return nil, nil
}
func (r *fakeTripRepo) ListActiveInPeriod(start, end time.Time) ([]models.Trip, error) {
	return nil, nil
}

var (
	_ user.UserRepository                          = (*fakeUserRepo)(nil)
	_ task.TaskRepository                          = (*fakeTaskRepo)(nil)
	_ project.ProjectRepository                    = (*fakeProjectRepo)(nil)
	_ status.StatusRepository                      = (*fakeStatusRepo)(nil)
	_ status_transition.StatusTransitionRepository = (*fakeStatusTransitionRepo)(nil)
	_ report.ReportRepository                      = (*fakeReportRepo)(nil)
	_ system.SystemRepository                      = (*fakeSystemRepo)(nil)
	_ participant.ParticipantRepository            = (*fakeParticipantRepo)(nil)
	_ trip.TripRepository                          = (*fakeTripRepo)(nil)
)

func decodeJSONBody[T any](t *testing.T, body *bytes.Buffer) T {
	t.Helper()

	var result T
	if err := json.Unmarshal(body.Bytes(), &result); err != nil {
		t.Fatalf("decode json body: %v", err)
	}
	return result
}
