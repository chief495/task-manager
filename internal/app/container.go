package app

import (
	"os"

	"github.com/eduardnesterov/task-manager/internal/organization"
	organization_contact "github.com/eduardnesterov/task-manager/internal/organization/contact"
	"github.com/eduardnesterov/task-manager/internal/platform/email"
	"github.com/eduardnesterov/task-manager/internal/platform/sheduler"
	"github.com/eduardnesterov/task-manager/internal/project"
	"github.com/eduardnesterov/task-manager/internal/project/labor"
	"github.com/eduardnesterov/task-manager/internal/project/participant"
	project_phase "github.com/eduardnesterov/task-manager/internal/project/phase"
	"github.com/eduardnesterov/task-manager/internal/project/system"
	"github.com/eduardnesterov/task-manager/internal/project/system/product"
	product_stage "github.com/eduardnesterov/task-manager/internal/project/system/product/stage"
	system_type "github.com/eduardnesterov/task-manager/internal/project/system/type"
	"github.com/eduardnesterov/task-manager/internal/project/trip"
	work_type "github.com/eduardnesterov/task-manager/internal/project/work_type"
	"github.com/eduardnesterov/task-manager/internal/report"
	"github.com/eduardnesterov/task-manager/internal/report/export"
	"github.com/eduardnesterov/task-manager/internal/task"
	task_template "github.com/eduardnesterov/task-manager/internal/task/template"
	"github.com/eduardnesterov/task-manager/internal/user"
	userauth "github.com/eduardnesterov/task-manager/internal/user/auth"
	"github.com/eduardnesterov/task-manager/internal/workflow/stage"
	stage_transition "github.com/eduardnesterov/task-manager/internal/workflow/stage/transition"
	"github.com/eduardnesterov/task-manager/internal/workflow/status"
	status_transition "github.com/eduardnesterov/task-manager/internal/workflow/status/transition"
	"gorm.io/gorm"
)

type appHandlers struct {
	auth                *userauth.AuthHandler
	project             *project.ProjectHandler
	projectPhase        *project_phase.ProjectPhaseHandler
	user                *user.UserHandler
	task                *task.TaskHandler
	status              *status.StatusHandler
	stage               *stage.StageHandler
	statusTransition    *status_transition.StatusTransitionHandler
	stageTransition     *stage_transition.StageTransitionHandler
	report              *report.ReportHandler
	system              *system.SystemHandler
	product             *product.ProductHandler
	workType            *work_type.WorkTypeHandler
	systemType          *system_type.SystemTypeHandler
	productStage        *product_stage.ProductStageHandler
	organization        *organization.OrganizationHandler
	organizationContact *organization_contact.OrganizationContactHandler
	participant         *participant.ParticipantHandler
	trip                *trip.TripHandler
	taskTemplate        *task_template.TaskTemplateHandler
	export              *export.ExportHandler
	labor               *labor.LaborHandler
}

func newAppHandlers(db *gorm.DB) (*appHandlers, *sheduler.Scheduler) {
	projectRepo := project.NewProjectRepo(db)
	projectPhaseRepo := project_phase.NewProjectPhaseRepo(db)
	userRepo := user.NewUserRepo(db)
	taskRepo := task.NewTaskRepo(db)
	statusRepo := status.NewStatusRepo(db)
	stageRepo := stage.NewStageRepo(db)
	statusTransitionRepo := status_transition.NewStatusTransitionRepo(db)
	stageTransitionRepo := stage_transition.NewStageTransitionRepo(db)
	reportRepo := report.NewReportRepo(db)
	systemRepo := system.NewSystemRepo(db)
	participantRepo := participant.NewParticipantRepo(db)
	tripRepo := trip.NewTripRepo(db)
	productRepo := product.NewProductRepo(db)
	workTypeRepo := work_type.NewWorkTypeRepo(db)
	systemTypeRepo := system_type.NewSystemTypeRepo(db)
	productStageRepo := product_stage.NewProductStageRepo(db)
	organizationRepo := organization.NewOrganizationRepo(db)
	organizationContactRepo := organization_contact.NewOrganizationContactRepo(db)
	templateRepo := task_template.NewTaskTemplateRepo(db)
	laborRepo := labor.NewLaborRepository(db)

	projectService := project.NewProjectService(projectRepo, userRepo, systemRepo, participantRepo, tripRepo)
	projectPhaseService := project_phase.NewProjectPhaseService(projectPhaseRepo, projectRepo, userRepo)
	emailService := email.NewEmailService(
		os.Getenv("SMTP_HOST"),
		os.Getenv("SMTP_PORT"),
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
		os.Getenv("SMTP_FROM"),
	)
	userService := user.NewUserService(userRepo, emailService)
	taskService := task.NewTaskService(taskRepo, statusRepo, statusTransitionRepo, userRepo, emailService)
	statusService := status.NewStatusService(statusRepo)
	stageService := stage.NewStageService(stageRepo)
	statusTransitionService := status_transition.NewStatusTransitionService(statusTransitionRepo)
	stageTransitionService := stage_transition.NewStageTransitionService(stageTransitionRepo)
	reportService := report.NewReportService(reportRepo)
	systemService := system.NewSystemService(systemRepo)
	productService := product.NewProductService(productRepo)
	workTypeService := work_type.NewWorkTypeService(workTypeRepo)
	systemTypeService := system_type.NewSystemTypeService(systemTypeRepo)
	productStageService := product_stage.NewProductStageService(productStageRepo)
	organizationService := organization.NewOrganizationService(organizationRepo)
	organizationContactService := organization_contact.NewOrganizationContactService(organizationContactRepo)
	participantService := participant.NewParticipantService(participantRepo)
	tripService := trip.NewTripService(tripRepo)
	templateService := task_template.NewTaskTemplateService(templateRepo)
	laborService := labor.NewLaborService(laborRepo, projectRepo)
	exportService := export.NewExportService(taskRepo, projectRepo, userRepo, reportRepo)
	dailyScheduler := sheduler.NewScheduler(userRepo)

	return &appHandlers{
		auth:                userauth.NewAuthHandler(userService),
		project:             project.NewProjectHandler(projectService),
		projectPhase:        project_phase.NewProjectPhaseHandler(projectPhaseService),
		user:                user.NewUserHandler(userService),
		task:                task.NewTaskHandler(taskService),
		status:              status.NewStatusHandler(statusService),
		stage:               stage.NewStageHandler(stageService),
		statusTransition:    status_transition.NewStatusTransitionHandler(statusTransitionService),
		stageTransition:     stage_transition.NewStageTransitionHandler(stageTransitionService),
		report:              report.NewReportHandler(reportService, projectService),
		system:              system.NewSystemHandler(systemService),
		product:             product.NewProductHandler(productService),
		workType:            work_type.NewWorkTypeHandler(workTypeService),
		systemType:          system_type.NewSystemTypeHandler(systemTypeService),
		productStage:        product_stage.NewProductStageHandler(productStageService),
		organization:        organization.NewOrganizationHandler(organizationService),
		organizationContact: organization_contact.NewOrganizationContactHandler(organizationContactService),
		participant:         participant.NewParticipantHandler(participantService),
		trip:                trip.NewTripHandler(tripService),
		taskTemplate:        task_template.NewTaskTemplateHandler(templateService),
		export:              export.NewExportHandler(exportService),
		labor:               labor.NewLaborHandler(laborService),
	}, dailyScheduler
}
