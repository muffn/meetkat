package poll

// Repository defines the persistence interface for polls.
type Repository interface {
	Create(p *Poll) error
	GetByPublicID(publicID string) (*Poll, error)
	GetByAdminID(adminID string) (*Poll, error)
	AddVote(pollID string, vote Vote) error
	RemoveVote(pollID string, voterName string) error
}
