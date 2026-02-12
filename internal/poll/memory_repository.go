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

func (r *MemoryRepository) GetByAdminID(adminID string) (*Poll, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.polls {
		if p.AdminID == adminID {
			return p, nil
		}
	}
	return nil, nil
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

func (r *MemoryRepository) RemoveVote(pollID string, voterName string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[pollID]
	if !ok {
		return errors.New("poll not found")
	}
	for i, v := range p.Votes {
		if v.Name == voterName {
			p.Votes = append(p.Votes[:i], p.Votes[i+1:]...)
			return nil
		}
	}
	return errors.New("vote not found")
}

func (r *MemoryRepository) Delete(pollID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.polls[pollID]; !ok {
		return errors.New("poll not found")
	}
	delete(r.polls, pollID)
	return nil
}

func (r *MemoryRepository) UpdateVote(pollID string, oldName string, vote Vote) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.polls[pollID]
	if !ok {
		return errors.New("poll not found")
	}
	for i, v := range p.Votes {
		if v.Name == oldName {
			p.Votes[i] = vote
			return nil
		}
	}
	return errors.New("vote not found")
}
