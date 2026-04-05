package member

import (
	"errors"
	"time"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(m Member) (Member, error) {
	if m.Name == "" {
		return Member{}, errors.New("name is required")
	}

	if m.Status == "" {
		m.Status = Active
	}

	if m.Status != Active && m.Status != Inactive {
		return Member{}, errors.New("invalid status")
	}

	now := time.Now()

	m.CreatedAt = now
	m.UpdatedAt = now

	result, err := s.repo.Create(m)
	if err != nil {
		return Member{}, err
	}

	return result, nil
}

type PaginationMeta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

func (s *Service) Find(filters map[string]string, page, limit int) ([]Member, PaginationMeta, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	data, err := s.repo.Find(filters, limit, offset)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	total, err := s.repo.Count(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	meta := PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}

	return data, meta, nil
}
