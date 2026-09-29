package mocks

//go:generate go run go.uber.org/mock/mockgen -destination=user_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/user UserRepository
//go:generate go run go.uber.org/mock/mockgen -destination=project_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project ProjectRepository
//go:generate go run go.uber.org/mock/mockgen -destination=task_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/task TaskRepository
//go:generate go run go.uber.org/mock/mockgen -destination=status_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/workflow/status StatusRepository
//go:generate go run go.uber.org/mock/mockgen -destination=status_transition_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/workflow/status/transition StatusTransitionRepository
//go:generate go run go.uber.org/mock/mockgen -destination=report_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/report ReportRepository
//go:generate go run go.uber.org/mock/mockgen -destination=system_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/system SystemRepository
//go:generate go run go.uber.org/mock/mockgen -destination=participant_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/participant ParticipantRepository
//go:generate go run go.uber.org/mock/mockgen -destination=trip_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/trip TripRepository
//go:generate go run go.uber.org/mock/mockgen -destination=stage_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/workflow/stage StageRepository
//go:generate go run go.uber.org/mock/mockgen -destination=stage_transition_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/workflow/stage/transition StageTransitionRepository
//go:generate go run go.uber.org/mock/mockgen -destination=product_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/system/product ProductRepository
//go:generate go run go.uber.org/mock/mockgen -destination=work_type_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/work_type WorkTypeRepository
//go:generate go run go.uber.org/mock/mockgen -destination=system_type_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/system/type SystemTypeRepository
//go:generate go run go.uber.org/mock/mockgen -destination=product_stage_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/system/product/stage ProductStageRepository
//go:generate go run go.uber.org/mock/mockgen -destination=organization_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/organization OrganizationRepository
//go:generate go run go.uber.org/mock/mockgen -destination=organization_contact_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/organization/contact OrganizationContactRepository
//go:generate go run go.uber.org/mock/mockgen -destination=project_phase_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/project/phase ProjectPhaseRepository
//go:generate go run go.uber.org/mock/mockgen -destination=task_template_repository_mocks.go -package=mocks github.com/eduardnesterov/task-manager/internal/task/template TaskTemplateRepository
