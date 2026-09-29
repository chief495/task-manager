package phase

type ProjectPhaseRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	ListByProject(projectID uint) ([]any, error)
}
