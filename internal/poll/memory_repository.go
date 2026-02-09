package poll

import (
	"errors"
	"sync"
)

// MemoryRepository is an in-memory implementation of Repository.
type MemoryRepository struct {
	mu    sync.Mutex
	polls map[string]*Poll
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		polls: make(map[string]*Poll),
	}
}

func (r *MemoryRepository) Create(p *Poll) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.polls[p.ID] = p
	return nil
}

func (r *MemoryRepository) GetByPublicID(publicID string) (*Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[publicID]
	if !ok {
		return nil, nil
	}
	return p, nil
}

func (r *MemoryRepository) AddVote(pollID string, vote Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[pollID]
	if !ok {
		return errors.New("poll not found")
	}
	p.Votes = append(p.Votes, vote)
	return nil
}
