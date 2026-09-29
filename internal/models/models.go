package models

import (
	contactmodel "github.com/eduardnesterov/task-manager/internal/organization/contact/model"
	organizationmodel "github.com/eduardnesterov/task-manager/internal/organization/model"
	labormodel "github.com/eduardnesterov/task-manager/internal/project/labor/model"
	projectmodel "github.com/eduardnesterov/task-manager/internal/project/model"
	participantmodel "github.com/eduardnesterov/task-manager/internal/project/participant/model"
	phasemodel "github.com/eduardnesterov/task-manager/internal/project/phase/model"
	systemmodel "github.com/eduardnesterov/task-manager/internal/project/system/model"
	productmodel "github.com/eduardnesterov/task-manager/internal/project/system/product/model"
	productstagemodel "github.com/eduardnesterov/task-manager/internal/project/system/product/stage/model"
	systemtypemodel "github.com/eduardnesterov/task-manager/internal/project/system/type/model"
	tripmodel "github.com/eduardnesterov/task-manager/internal/project/trip/model"
	worktypemodel "github.com/eduardnesterov/task-manager/internal/project/work_type/model"
	reportmodel "github.com/eduardnesterov/task-manager/internal/report/model"
	taskmodel "github.com/eduardnesterov/task-manager/internal/task/model"
	templatemodel "github.com/eduardnesterov/task-manager/internal/task/template/model"
	usermodel "github.com/eduardnesterov/task-manager/internal/user/model"
	stagemodel "github.com/eduardnesterov/task-manager/internal/workflow/stage/model"
	stagetransitionmodel "github.com/eduardnesterov/task-manager/internal/workflow/stage/transition/model"
	statusmodel "github.com/eduardnesterov/task-manager/internal/workflow/status/model"
	statustransitionmodel "github.com/eduardnesterov/task-manager/internal/workflow/status/transition/model"
)

type User = usermodel.User
type TaskStats = usermodel.TaskStats
type SocialRatingLog = usermodel.SocialRatingLog

type Project = projectmodel.Project
type ProjectMember = projectmodel.ProjectMember
type ProjectAccess = projectmodel.ProjectAccess
type Stakeholder = projectmodel.Stakeholder
type ProjectMessage = projectmodel.ProjectMessage

type ProjectPhase = phasemodel.ProjectPhase
type Participant = participantmodel.Participant
type Trip = tripmodel.Trip
type WorkType = worktypemodel.WorkType

type System = systemmodel.System
type SystemType = systemtypemodel.SystemType
type Product = productmodel.Product
type ProductStage = productstagemodel.ProductStage

type Organization = organizationmodel.Organization
type OrganizationContact = contactmodel.OrganizationContact

type Stage = stagemodel.Stage
type StageTransition = stagetransitionmodel.StageTransition
type Status = statusmodel.Status
type StatusTransition = statustransitionmodel.StatusTransition

type Task = taskmodel.Task
type TaskTemplate = templatemodel.TaskTemplate
type Report = reportmodel.Report

type LaborCategory = labormodel.LaborCategory
type LaborNorm = labormodel.LaborNorm
type LaborCoefficient = labormodel.LaborCoefficient
type LaborCalculation = labormodel.LaborCalculation
type LaborCalculationEntry = labormodel.LaborCalculationEntry
type LaborCalculationEntryCoeff = labormodel.LaborCalculationEntryCoeff
type LaborCalculationGlobalCoeff = labormodel.LaborCalculationGlobalCoeff
type LaborCalculationRequest = labormodel.LaborCalculationRequest
type LaborCalculationItem = labormodel.LaborCalculationItem
type LaborItemCoefficient = labormodel.LaborItemCoefficient
type LaborGlobalCoefficient = labormodel.LaborGlobalCoefficient
type LaborCalculationResult = labormodel.LaborCalculationResult
type LaborCalculationResultItem = labormodel.LaborCalculationResultItem
type LaborCategoryTotal = labormodel.LaborCategoryTotal
type LaborCalculationSummary = labormodel.LaborCalculationSummary
