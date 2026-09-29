package task

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/platform/email"
	"github.com/eduardnesterov/task-manager/internal/user"
	"github.com/eduardnesterov/task-manager/internal/workflow/status"
	status_transition "github.com/eduardnesterov/task-manager/internal/workflow/status/transition"
)

// Ошибки
var (
	ErrTaskNotFound            = errors.New("задача не найдена")
	ErrInvalidStatusTransition = errors.New("недопустимый переход статуса")
)

type TaskService struct {
	repo                 TaskRepository
	statusRepo           status.StatusRepository
	statusTransitionRepo status_transition.StatusTransitionRepository
	userRepo             user.UserRepository
	emailService         *email.EmailService
}

func NewTaskService(repo TaskRepository, statusRepo status.StatusRepository, statusTransitionRepo status_transition.StatusTransitionRepository, userRepo user.UserRepository, emailService *email.EmailService) *TaskService {
	return &TaskService{
		repo:                 repo,
		statusRepo:           statusRepo,
		statusTransitionRepo: statusTransitionRepo,
		userRepo:             userRepo,
		emailService:         emailService,
	}
}

// CreateTask — создание новой задачи
func (s *TaskService) CreateTask(task *models.Task, assignerID uint) error {
	if task.StartTime == nil || task.EndTime == nil {
		return fmt.Errorf("start_time and end_time are required")
	}
	if task.Name == "" {
		return errors.New("название задачи не может быть пустым")
	}
	if err := s.validateTaskPhase(task); err != nil {
		return err
	}

	assignerName := ""
	if assignerID != 0 {
		assigner, err := s.userRepo.GetByID(assignerID)
		if err == nil {
			if assignerModel, ok := assigner.(*models.User); ok {
				assignerName = assignerModel.Name
			}
		}
	}

	user, err := s.userRepo.GetByID(task.ExecutorID)
	if err != nil {
		return err
	}
	planTime := calculateWorkingHours(*task.StartTime, *task.EndTime)
	userModel := user.(*models.User)
	if userModel.MiddlePlanTime > 0 {
		userModel.MiddlePlanTime = float32(planTime)*0.2 + userModel.MiddlePlanTime*0.8
	} else {
		userModel.MiddlePlanTime = float32(planTime)
	}
	err = s.userRepo.Update(userModel)
	if err != nil {
		return err
	}
	if err = s.repo.Create(task); err != nil {
		return err
	}

	if s.emailService != nil && userModel.Email != "" {
		if err = s.emailService.SendTaskAssigned(
			userModel.Email,
			userModel.Name,
			assignerName,
			task.Name,
			task.Description,
			task.StartTime,
			task.EndTime,
			task.Priority,
			task.ID,
		); err != nil {
			log.Printf("failed to send task assignment email for task_id=%d executor_id=%d: %v", task.ID, task.ExecutorID, err)
		}
	}

	return nil
}

// GetTask — получение задачи по ID с загрузкой связей
func (s *TaskService) GetTask(id uint) (*models.Task, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Task), nil
}

// UpdateTask — обновление задачи
func (s *TaskService) UpdateTask(task *models.Task) error {
	if task.ID == 0 {
		return errors.New("ID задачи не указан")
	}
	statusEntity, err := s.statusRepo.GetByID(task.StatusID)
	if err == nil && statusEntity.(*models.Status).IsFinal {
		task.Completed = true
	}
	if err := s.validateTaskPhase(task); err != nil {
		return err
	}
	return s.repo.Update(task)
}

// DeleteTask — удаление задачи
func (s *TaskService) DeleteTask(id uint) error {
	return s.repo.Delete(id)
}

// ListTasks — получение списка всех задач
func (s *TaskService) ListTasks(userId uint) ([]*models.Task, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}
	userModel, err := s.userRepo.GetByID(userId)
	if err != nil {
		return nil, err
	}
	user := userModel.(*models.User)
	var tasks []*models.Task
	for _, e := range entities {
		eModel := e.(*models.Task)
		if eModel.ExecutorID == user.ID || user.Role == "admin" {
			tasks = append(tasks, eModel)
		}
	}
	return tasks, nil
}

// CountTasks — подсчёт количества задач
func (s *TaskService) CountTasks() (int64, error) {
	return s.repo.Count()
}

// ListTasksByProject — задачи по проекту
func (s *TaskService) ListTasksByProject(projectID uint) ([]*models.Task, error) {
	entities, err := s.repo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*models.Task, len(entities))
	for i, e := range entities {
		tasks[i] = e.(*models.Task)
	}
	return tasks, nil
}

// AssignExecutor — назначение исполнителя на задачу
func (s *TaskService) AssignExecutor(taskID, executorID uint) error {
	task, err := s.GetTask(taskID)
	if err != nil {
		return err
	}

	task.ExecutorID = executorID
	return s.UpdateTask(task)
}

// UpdateTaskStatus — изменение статуса задачи с проверкой переходов
func (s *TaskService) UpdateTaskStatus(taskID, newStatusID uint) (*models.Task, error) {
	// Получаем задачу
	task, err := s.GetTask(taskID)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	// Получаем текущий статус
	currentStatus, err := s.statusRepo.GetByID(task.StatusID)
	if err != nil {
		return nil, err
	}

	// Проверяем, есть ли разрешённый переход
	allowed, err := s.statusTransitionRepo.ExistsTransition(currentStatus.(*models.Status).ID, newStatusID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrInvalidStatusTransition
	}
	// Обновляем статус
	task.StatusID = newStatusID
	statusName := ""
	statusIsFinal := false
	newStatus, err := s.statusRepo.GetByID(newStatusID)
	if err == nil {
		newStatusModel := newStatus.(*models.Status)
		statusName = newStatusModel.Name
		statusIsFinal = newStatusModel.IsFinal
	}
	if statusName == "На проверке" {
		now := time.Now()
		task.FactEndTime = &now
	}
	if statusName == "Пауза" {
		now := time.Now()
		task.PauseDate = &now
	}
	if currentStatus.(*models.Status).Name == "Пауза" && statusName == "В работе" {
		oldEndTime := *task.EndTime
		newEndTime := oldEndTime.Add(time.Now().Sub(*task.PauseDate))
		task.EndTime = &newEndTime
	}
	task.Completed = statusIsFinal
	if statusIsFinal {
		if task.StartTime != nil && task.EndTime != nil && task.FactEndTime != nil && statusName != "Аннулировано" {
			// Рассчитываем фактическое время только по рабочим часам
			task.FactTime = float32(calculateWorkingHours(*task.StartTime, *task.FactEndTime))

			user, err := s.userRepo.GetByID(task.ExecutorID)
			if err != nil {
				return nil, err
			}
			userModel := user.(*models.User)

			// Плановое время в рабочих часах
			planTimeHours := calculateWorkingHours(*task.StartTime, *task.EndTime)

			// Защита от деления на ноль
			if task.FactTime == 0 || userModel.MiddlePlanTime == 0 {
				return nil, fmt.Errorf("некорректное время выполнения или среднее плановое время")
			}

			// Коэффициент эффективности
			coef := task.ComplexityCoefficient * task.UrgencyCoefficient * planTimeHours / task.FactTime

			// Обновление социального рейтинга
			userModel.SocialRating = 0.8*userModel.SocialRating + 0.2*coef

			err = s.userRepo.Update(userModel)
			if err != nil {
				return nil, err
			}
		}
	}
	if err := s.repo.Update(task); err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskService) ListTasksByCategory(category string) ([]*models.Task, error) {
	entities, err := s.repo.ListByCategory(category)
	if err != nil {
		return nil, err
	}

	tasks := make([]*models.Task, len(entities))
	for i, e := range entities {
		tasks[i] = e.(*models.Task)
	}
	return tasks, nil
}

func (s *TaskService) ListTasksByPhase(phaseID uint) ([]*models.Task, error) {
	entities, err := s.repo.ListByPhase(phaseID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*models.Task, len(entities))
	for i, e := range entities {
		tasks[i] = e.(*models.Task)
	}
	return tasks, nil
}

func (s *TaskService) validateTaskPhase(task *models.Task) error {
	if task.ProjectPhaseID == nil {
		return nil
	}

	var count int64
	err := s.repo.GetDB().
		Model(&models.ProjectPhase{}).
		Where("id = ? AND project_id = ?", *task.ProjectPhaseID, task.ProjectID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New("этап не относится к указанному проекту")
	}
	return nil
}
