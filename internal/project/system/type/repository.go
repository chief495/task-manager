package systemtype

type SystemTypeRepository interface {
	Create(entity any) error
	GetByID(id uint) (any, error)
	Update(entity any) error
	Delete(id uint) error
	List() ([]any, error)
	GetByName(name string) (any, error)
	Search(query string) ([]any, error)
}
