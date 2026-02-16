package poll

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"
)

const idChars = "abcdefghijklmnopqrstuvwxyz0123456789"

type Vote struct {
	Name      string
	Responses map[string]string // key = option string, value = "yes", "no", "maybe", or ""
}

type Poll struct {
	ID          string
	AdminID     string
	Title       string
	Description string
	AnswerMode  string // "yn" (yes/no) or "ymn" (yes/maybe/no); default "yn"
	Options     []string
	Votes       []Vote
	CreatedAt   time.Time
}

type OptionTotal struct {
	Yes   int
	Maybe int
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func generateID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	for i := range b {
		b[i] = idChars[b[i]%byte(len(idChars))]
	}
	return string(b)
}

const (
	MaxTitleLen       = 200
	MaxDescriptionLen = 2000
	MaxNameLen        = 100
	MaxOptions        = 60
)

func (s *Service) Create(title, description, answerMode string, options []string) (*Poll, error) {
	if len(title) > MaxTitleLen {
		return nil, fmt.Errorf("title exceeds %d characters", MaxTitleLen)
	}
	if len(description) > MaxDescriptionLen {
		return nil, fmt.Errorf("description exceeds %d characters", MaxDescriptionLen)
	}
	if len(options) > MaxOptions {
		return nil, fmt.Errorf("too many options (max %d)", MaxOptions)
	}
	if answerMode != "yn" && answerMode != "ymn" {
		answerMode = "yn"
	}

	p := &Poll{
		ID:          generateID(),
		AdminID:     generateID(),
		Title:       title,
		Description: description,
		AnswerMode:  answerMode,
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

func (s *Service) GetByAdminID(adminID string) (*Poll, error) {
	return s.repo.GetByAdminID(adminID)
}

func (s *Service) RemoveVote(pollID, voterName string) error {
	return s.repo.RemoveVote(pollID, voterName)
}

func (s *Service) AddVote(pollID, name string, responses map[string]string) error {
	if name == "" {
		return errors.New("name must not be empty")
	}
	if len(name) > MaxNameLen {
		return fmt.Errorf("name exceeds %d characters", MaxNameLen)
	}
	return s.repo.AddVote(pollID, Vote{Name: name, Responses: responses})
}

func (s *Service) Delete(pollID string) error {
	return s.repo.Delete(pollID)
}

func (s *Service) UpdateVote(pollID, oldName, newName string, responses map[string]string) error {
	if newName == "" {
		return errors.New("name must not be empty")
	}
	if len(newName) > MaxNameLen {
		return fmt.Errorf("name exceeds %d characters", MaxNameLen)
	}
	return s.repo.UpdateVote(pollID, oldName, Vote{Name: newName, Responses: responses})
}

func Totals(p *Poll) map[string]OptionTotal {
	totals := make(map[string]OptionTotal, len(p.Options))
	for _, opt := range p.Options {
		var t OptionTotal
		for _, v := range p.Votes {
			switch v.Responses[opt] {
			case "yes":
				t.Yes++
			case "maybe":
				t.Maybe++
			}
		}
		totals[opt] = t
	}
	return totals
}
