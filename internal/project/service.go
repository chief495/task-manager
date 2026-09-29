package project

import (
	"errors"
	"strings"

	"github.com/eduardnesterov/task-manager/internal/models"
	"github.com/eduardnesterov/task-manager/internal/project/participant"
	"github.com/eduardnesterov/task-manager/internal/project/system"
	"github.com/eduardnesterov/task-manager/internal/project/trip"
	"github.com/eduardnesterov/task-manager/internal/user"
	stage_transition "github.com/eduardnesterov/task-manager/internal/workflow/stage/transition"
	"gorm.io/gorm"
)

// ProjectService — сервис для работы с проектами
type ProjectService struct {
	repo            ProjectRepository
	userRepo        user.UserRepository
	systemRepo      system.SystemRepository
	participantRepo participant.ParticipantRepository
	tripRepo        trip.TripRepository
}

// NewProjectService — конструктор сервиса
func NewProjectService(
	repo ProjectRepository,
	userRepo user.UserRepository,
	systemRepo system.SystemRepository,
	participantRepo participant.ParticipantRepository,
	tripRepo trip.TripRepository,
) *ProjectService {
	return &ProjectService{
		repo:            repo,
		userRepo:        userRepo,
		systemRepo:      systemRepo,
		participantRepo: participantRepo,
		tripRepo:        tripRepo,
	}
}

// CreateProject — создание нового проекта
func (s *ProjectService) CreateProject(project *models.Project) error {
	if project.Name == "" {
		return errors.New("название проекта не может быть пустым")
	}
	if project.StageID == 0 {
		return errors.New("стадия проекта не указана")
	}

	return s.repo.Create(project)
}

func (s *ProjectService) GetDB() *gorm.DB {
	return s.repo.GetDB()
}

// GetProject — получение проекта по ID с загрузкой связей
func (s *ProjectService) GetProject(id uint) (*models.Project, error) {
	entity, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Project), nil
}

func (s *ProjectService) UserCanAccessProject(projectID, userID uint, role string) (bool, error) {
	if role == "admin" {
		return true, nil
	}

	var project models.Project
	err := s.repo.GetDB().
		Select("id", "is_hidden", "responsible_id").
		First(&project, projectID).Error
	if err != nil {
		return false, err
	}
	var count int64
	if project.ResponsibleID == userID {
		return true, nil
	}

	err = s.repo.GetDB().
		Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}

	err = s.repo.GetDB().
		Model(&models.Task{}).
		Where("project_id = ? AND executor_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}

	if project.IsHidden {
		err = s.repo.GetDB().
			Model(&models.ProjectAccess{}).
			Where("project_id = ? AND user_id = ?", projectID, userID).
			Count(&count).Error
	}
	return count > 0, err
}

// GetTasksByProjectIDs — получение всех задач по списку ID проектов
func (s *ProjectService) GetTasksByProjectIDs(projectIDs []uint) ([]models.Task, error) {
	var tasks []models.Task
	if len(projectIDs) == 0 {
		return tasks, nil
	}

	err := s.repo.GetDB().Model(&models.Task{}).
		Where("project_id IN ?", projectIDs).
		Find(&tasks).Error

	return tasks, err
}

// UpdateProject — обновление проекта
func (s *ProjectService) UpdateProject(project *models.Project) error {
	if project.ID == 0 {
		return errors.New("ID проекта не указан")
	}
	if project.Name == "" {
		return errors.New("название проекта не может быть пустым")
	}
	if _, err := s.userRepo.GetByID(project.ResponsibleID); err != nil {
		return err
	}
	return s.repo.Update(project)
}

// DeleteProject — удаление проекта
func (s *ProjectService) DeleteProject(id uint) error {
	return s.repo.Delete(id)
}

func (s *ProjectService) CreateStakeholder(stakeholder *models.Stakeholder) error {
	if stakeholder.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if stakeholder.OrganizationID == 0 {
		return errors.New("organization_id не указан")
	}
	if stakeholder.FullName == "" {
		return errors.New("ФИО заинтересованного лица не указано")
	}
	return s.repo.GetDB().Create(stakeholder).Error
}

func (s *ProjectService) UpdateStakeholder(stakeholder *models.Stakeholder) error {
	if stakeholder.ID == 0 {
		return errors.New("ID заинтересованного лица не указан")
	}
	if stakeholder.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if stakeholder.OrganizationID == 0 {
		return errors.New("organization_id не указан")
	}
	if stakeholder.FullName == "" {
		return errors.New("ФИО заинтересованного лица не указано")
	}
	return s.repo.GetDB().Save(stakeholder).Error
}

// ListProjects — получение списка всех проектов
func (s *ProjectService) ListProjects() ([]*models.Project, error) {
	entities, err := s.repo.List()
	if err != nil {
		return nil, err
	}

	projects := make([]*models.Project, len(entities))
	for i, e := range entities {
		projects[i] = e.(*models.Project)
	}
	return projects, nil
}

func (s *ProjectService) ListProjectsForUser(userID uint, role string) ([]*models.Project, error) {
	if role == "admin" {
		return s.ListProjects()
	}

	projects, err := s.repo.ListVisibleForUser(userID)
	if err != nil {
		return nil, err
	}
	result := make([]*models.Project, len(projects))
	for i := range projects {
		result[i] = &projects[i]
	}
	return result, nil
}

func (s *ProjectService) ReplaceProjectAccess(projectID uint, userIDs []uint) error {
	return s.repo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Where("project_id = ?", projectID).Delete(&models.ProjectAccess{}).Error; err != nil {
			return err
		}

		seen := make(map[uint]bool, len(userIDs))
		for _, userID := range userIDs {
			if userID == 0 || seen[userID] {
				continue
			}
			seen[userID] = true
			if err := tx.Create(&models.ProjectAccess{ProjectID: projectID, UserID: userID}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *ProjectService) ListMessages(projectID uint) ([]models.ProjectMessage, error) {
	if projectID == 0 {
		return nil, errors.New("project_id не указан")
	}

	var messages []models.ProjectMessage
	err := s.repo.GetDB().
		Where("project_id = ?", projectID).
		Preload("User").
		Preload("ReplyTo").
		Preload("ReplyTo.User").
		Order("created_at ASC, id ASC").
		Find(&messages).Error
	return messages, err
}

func (s *ProjectService) CreateMessage(message *models.ProjectMessage) error {
	message.Content = strings.TrimSpace(message.Content)
	if message.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if message.UserID == 0 {
		return errors.New("user_id не указан")
	}
	if message.Content == "" {
		return errors.New("сообщение не может быть пустым")
	}
	if _, err := s.repo.GetByID(message.ProjectID); err != nil {
		return err
	}
	if _, err := s.userRepo.GetByID(message.UserID); err != nil {
		return err
	}
	if message.ReplyToID != nil {
		replyTo, err := s.GetMessage(*message.ReplyToID)
		if err != nil {
			return err
		}
		if replyTo.ProjectID != message.ProjectID {
			return errors.New("нельзя ответить на сообщение из другого проекта")
		}
	}
	return s.repo.GetDB().Create(message).Error
}

func (s *ProjectService) GetMessage(id uint) (*models.ProjectMessage, error) {
	var message models.ProjectMessage
	err := s.repo.GetDB().
		Preload("User").
		Preload("ReplyTo").
		Preload("ReplyTo.User").
		First(&message, id).Error
	return &message, err
}

func (s *ProjectService) UpdateMessage(message *models.ProjectMessage) error {
	message.Content = strings.TrimSpace(message.Content)
	if message.ID == 0 {
		return errors.New("ID сообщения не указан")
	}
	if message.Content == "" {
		return errors.New("сообщение не может быть пустым")
	}
	return s.repo.GetDB().Save(message).Error
}

func (s *ProjectService) DeleteMessage(id uint) error {
	return s.repo.GetDB().Delete(&models.ProjectMessage{}, id).Error
}

func (s *ProjectService) AddMember(member *models.ProjectMember) error {
	if member.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if member.UserID == 0 {
		return errors.New("user_id не указан")
	}
	if member.Role == "" {
		member.Role = "member"
	}
	if _, err := s.repo.GetByID(member.ProjectID); err != nil {
		return err
	}
	if _, err := s.userRepo.GetByID(member.UserID); err != nil {
		return err
	}
	return s.repo.CreateMember(member)
}

func (s *ProjectService) GetMember(id uint) (*models.ProjectMember, error) {
	return s.repo.GetMemberByID(id)
}

func (s *ProjectService) UpdateMember(member *models.ProjectMember) error {
	if member.ID == 0 {
		return errors.New("ID участника проекта не указан")
	}
	if member.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if member.UserID == 0 {
		return errors.New("user_id не указан")
	}
	if member.Role == "" {
		member.Role = "member"
	}
	if _, err := s.repo.GetByID(member.ProjectID); err != nil {
		return err
	}
	if _, err := s.userRepo.GetByID(member.UserID); err != nil {
		return err
	}
	return s.repo.UpdateMember(member)
}

func (s *ProjectService) DeleteMember(id uint) error {
	return s.repo.DeleteMember(id)
}

func (s *ProjectService) ListMembersByProject(projectID uint) ([]models.ProjectMember, error) {
	if projectID == 0 {
		return nil, errors.New("project_id не указан")
	}
	return s.repo.ListMembersByProject(projectID)
}

// CountProjects — подсчёт количества проектов
func (s *ProjectService) CountProjects() (int64, error) {
	return s.repo.Count()
}

// ListProjectsByStage — получение проектов по стадии
func (s *ProjectService) ListProjectsByStage(stageID uint) ([]models.Project, error) {
	projects, err := s.repo.ListByStage(stageID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// ListProjectsByResponsible — получение проектов по ответственному
func (s *ProjectService) ListProjectsByResponsible(responsibleID uint) ([]models.Project, error) {
	projects, err := s.repo.ListByResponsible(responsibleID)
	if err != nil {
		return nil, err
	}
	return projects, nil
}

// AssignManager — назначение менеджера проекта
func (s *ProjectService) AssignManager(projectID, managerID uint) error {
	project, err := s.GetProject(projectID)
	if err != nil {
		return err
	}

	project.ResponsibleID = managerID
	return s.UpdateProject(project)
}

// ChangeStage — смена стадии проекта (с проверкой переходов)
func (s *ProjectService) ChangeStage(
	projectID uint,
	toStageID uint,
	userRole string,
	stageTransitionService *stage_transition.StageTransitionService,
) error {
	project, err := s.GetProject(projectID)
	if err != nil {
		return err
	}

	// Проверяем возможность перехода
	can, msg, err := stageTransitionService.CanTransition(project, toStageID, userRole)
	if err != nil {
		return err
	}
	if !can {
		return errors.New(msg)
	}

	// Меняем стадию
	project.StageID = toStageID
	return s.UpdateProject(project)
}

// SearchProjects — поиск проектов по названию
func (s *ProjectService) SearchProjects(query string) ([]*models.Project, error) {
	entities, err := s.repo.Search(query)
	if err != nil {
		return nil, err
	}

	projects := make([]*models.Project, len(entities))
	for i, e := range entities {
		projects[i] = e.(*models.Project)
	}
	return projects, nil
}

// --- Системы проекта ---

// ListSystems — список систем проекта
func (s *ProjectService) ListSystems(projectID uint) ([]*models.System, error) {
	entities, err := s.systemRepo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.System, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.System)
	}
	return list, nil
}

// GetSystem — получение системы по ID
func (s *ProjectService) GetSystem(id uint) (*models.System, error) {
	entity, err := s.systemRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.System), nil
}

// CreateSystem — создание системы в проекте
func (s *ProjectService) CreateSystem(system *models.System) error {
	if system.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if system.Name == "" {
		return errors.New("название системы не может быть пустым")
	}
	return s.systemRepo.Create(system)
}

// UpdateSystem — обновление системы
func (s *ProjectService) UpdateSystem(system *models.System) error {
	if system.ID == 0 {
		return errors.New("ID системы не указан")
	}
	return s.systemRepo.Update(system)
}

// DeleteSystem — удаление системы
func (s *ProjectService) DeleteSystem(id uint) error {
	return s.systemRepo.Delete(id)
}

// --- Участники проекта ---

// ListParticipants — список участников проекта
func (s *ProjectService) ListParticipants(projectID uint) ([]*models.Participant, error) {
	entities, err := s.participantRepo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.Participant, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Participant)
	}
	return list, nil
}

// GetParticipant — получение участника по ID
func (s *ProjectService) GetParticipant(id uint) (*models.Participant, error) {
	entity, err := s.participantRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Participant), nil
}

// CreateParticipant — добавление участника к проекту
func (s *ProjectService) CreateParticipant(p *models.Participant) error {
	if p.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if p.OrganizationID == 0 {
		return errors.New("organization_id не указан")
	}
	return s.participantRepo.Create(p)
}

// UpdateParticipant — обновление участника
func (s *ProjectService) UpdateParticipant(p *models.Participant) error {
	if p.ID == 0 {
		return errors.New("ID участника не указан")
	}
	return s.participantRepo.Update(p)
}

// DeleteParticipant — удаление участника из проекта
func (s *ProjectService) DeleteParticipant(id uint) error {
	return s.participantRepo.Delete(id)
}

// --- Командировки проекта ---

// ListTrips — список командировок по проекту
func (s *ProjectService) ListTrips(projectID uint) ([]*models.Trip, error) {
	entities, err := s.tripRepo.ListByProject(projectID)
	if err != nil {
		return nil, err
	}
	list := make([]*models.Trip, len(entities))
	for i, e := range entities {
		list[i] = e.(*models.Trip)
	}
	return list, nil
}

// GetTrip — получение командировки по ID
func (s *ProjectService) GetTrip(id uint) (*models.Trip, error) {
	entity, err := s.tripRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return entity.(*models.Trip), nil
}

// CreateTrip — создание командировки
func (s *ProjectService) CreateTrip(trip *models.Trip) error {
	if trip.ProjectID == 0 {
		return errors.New("project_id не указан")
	}
	if trip.PersonID == 0 {
		return errors.New("person_id не указан")
	}
	return s.tripRepo.Create(trip)
}

// UpdateTrip — обновление командировки
func (s *ProjectService) UpdateTrip(trip *models.Trip) error {
	if trip.ID == 0 {
		return errors.New("ID командировки не указан")
	}
	return s.tripRepo.Update(trip)
}

// DeleteTrip — удаление командировки
func (s *ProjectService) DeleteTrip(id uint) error {
	return s.tripRepo.Delete(id)
}

// --- Типы работ проекта (many2many) ---

// ListWorkTypesForProject — список типов работ, привязанных к проекту
func (s *ProjectService) ListWorkTypesForProject(projectID uint) ([]models.WorkType, error) {
	var workTypes []models.WorkType
	err := s.repo.GetDB().
		Table("work_types").
		Joins("JOIN project_work_types ON project_work_types.work_type_id = work_types.id").
		Where("project_work_types.project_id = ?", projectID).
		Find(&workTypes).Error
	if err != nil {
		return nil, err
	}
	return workTypes, nil
}

// AddWorkTypeToProject — привязка типа работ к проекту
func (s *ProjectService) AddWorkTypeToProject(projectID, workTypeID uint) error {
	return s.repo.GetDB().Exec(
		"INSERT INTO project_work_types (project_id, work_type_id) VALUES (?, ?) ON CONFLICT DO NOTHING",
		projectID,
		workTypeID,
	).Error
}

// RemoveWorkTypeFromProject — отвязка типа работ от проекта
func (s *ProjectService) RemoveWorkTypeFromProject(projectID, workTypeID uint) error {
	return s.repo.GetDB().
		Exec("DELETE FROM project_work_types WHERE project_id = ? AND work_type_id = ?", projectID, workTypeID).
		Error
}
