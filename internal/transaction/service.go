package transaction

type service struct {
	repo Repository
}

// NewService creates a new transaction service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// Implement the Service interface methods here