package poll

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

type Vote struct {
	Name      string
	Responses map[string]bool // key = option string, value = available
}

type Poll struct {
	ID          string
	Title       string
	Description string
	Options     []string
	Votes       []Vote
	CreatedAt   time.Time
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func generateID() string {
	b := make([]byte, 8)
	for i := range b {
		b[i] = idChars[rand.IntN(len(idChars))]
	}
	return string(b)
}

func (s *Service) Create(title, description string, options []string) (*Poll, error) {
	p := &Poll{
		ID:          generateID(),
		Title:       title,
		Description: description,
		Options:     options,
		CreatedAt:   time.Now(),
	}
	if err := s.repo.Create(p); err != nil {
		return nil, fmt.Errorf("create poll: %w", err)
	}
	return p, nil
}

func (s *Service) Get(id string) (*Poll, error) {
	return s.repo.GetByPublicID(id)
}

func (s *Service) AddVote(pollID, name string, responses map[string]bool) error {
	if name == "" {
		return errors.New("name must not be empty")
	}
	return s.repo.AddVote(pollID, Vote{Name: name, Responses: responses})
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
