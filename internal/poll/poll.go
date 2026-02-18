package poll

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"time"
)

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

// generateID returns a 26-character lowercase base32 string with 128-bit entropy.
// 16 random bytes → base32 (no padding) → 26 chars, zero modulo bias.
func generateID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}
	return strings.ToLower(
		base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b),
	), nil
}

const (
	AnswerModeYN  = "yn"
	AnswerModeYMN = "ymn"

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
	if answerMode != AnswerModeYN && answerMode != AnswerModeYMN {
		answerMode = AnswerModeYN
	}

	id, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("generate poll id: %w", err)
	}
	adminID, err := generateID()
	if err != nil {
		return nil, fmt.Errorf("generate admin id: %w", err)
	}

	p := &Poll{
		ID:          id,
		AdminID:     adminID,
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
