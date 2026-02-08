package poll

import (
	"errors"
	"math/rand/v2"
	"sync"
	"time"
)

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

type Vote struct {
	Name      string
	Responses map[string]bool // key = option string, value = available
}

type Poll struct {
	ID        string
	Title     string
	Options   []string
	Votes     []Vote
	CreatedAt time.Time
}

type Service struct {
	mu    sync.Mutex
	polls map[string]*Poll
}

func NewService() *Service {
	return &Service{
		polls: make(map[string]*Poll),
	}
}

func generateID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = idChars[rand.IntN(len(idChars))]
	}
	return string(b)
}

func (s *Service) Create(title string, options []string) *Poll {
	id := generateID()
	p := &Poll{
		ID:        id,
		Title:     title,
		Options:   options,
		CreatedAt: time.Now(),
	}
	s.mu.Lock()
	s.polls[id] = p
	s.mu.Unlock()
	return p
}

func (s *Service) Get(id string) (*Poll, bool) {
	s.mu.Lock()
	p, ok := s.polls[id]
	s.mu.Unlock()
	return p, ok
}

func (s *Service) AddVote(pollID, name string, responses map[string]bool) error {
	if name == "" {
		return errors.New("name must not be empty")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.polls[pollID]
	if !ok {
		return errors.New("poll not found")
	}
	p.Votes = append(p.Votes, Vote{Name: name, Responses: responses})
	return nil
}

func Totals(p *Poll) map[string]int {
	totals := make(map[string]int, len(p.Options))
	for _, opt := range p.Options {
		for _, v := range p.Votes {
			if v.Responses[opt] {
				totals[opt]++
			}
		}
	}
	return totals
}
