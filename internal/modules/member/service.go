package member

import (
	"encoding/csv"
	"errors"
	"io"
	"strconv"
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

func (s *Service) Update(id int64, payload Member) (Member, error) {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return Member{}, errors.New("member not found")
	}

	if payload.Name != "" {
		member.Name = payload.Name
	}

	if payload.Email != "" {
		member.Email = payload.Email
	}

	if payload.Phone != "" {
		member.Phone = payload.Phone
	}

	if payload.Status != "" {
		member.Status = payload.Status
	}

	if payload.ChurchRole != "" {
		member.ChurchRole = payload.ChurchRole
	}

	if payload.MaritalStatus != "" {
		member.MaritalStatus = payload.MaritalStatus
	}

	if payload.OriginDenomination != "" {
		member.OriginDenomination = payload.OriginDenomination
	}

	member.UpdatedAt = time.Now()

	return s.repo.Update(member)
}

func (s *Service) SoftDelete(id int64) error {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return errors.New("member not found")
	}

	if member.Status == Inactive {
		return errors.New("member already inactive")
	}

	return s.repo.SoftDelete(id)
}

func (s *Service) ImportCSV(file io.Reader, userID int64) (ImportResult, error) {
	reader := csv.NewReader(file)

	rows, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, err
	}

	var members []Member
	result := ImportResult{}

	for index, row := range rows {
		if index == 0 {
			continue
		}

		/*
			CSV expected:

			name,
			email,
			phone,
			status,
			member_since,
			baptized,
			baptism_date,
			church_role,
			marital_status,
			origin_denomination,
			membership_course_completed,
			membership_course_completed_at,
			contacted,
			contacted_at
		*/

		if len(row) < 14 {
			result.Failed++
			result.Errors = append(
				result.Errors,
				"invalid row structure",
			)
			continue
		}

		baptized := parseOptionalBool(row[5])
		membershipCourseCompleted := parseOptionalBool(row[10])
		contacted := parseOptionalBool(row[12])

		memberSince := parseOptionalDate(row[4])
		baptismDate := parseOptionalDate(row[6])
		membershipCourseCompletedAt := parseOptionalDate(row[11])
		contactedAt := parseOptionalDate(row[13])

		member := Member{
			Name:                        row[0],
			Email:                       row[1],
			Phone:                       row[2],
			Status:                      Status(row[3]),
			MemberSince:                 memberSince,
			Baptized:                    baptized,
			BaptismDate:                 baptismDate,
			ChurchRole:                  row[7],
			MaritalStatus:               MaritalStatus(row[8]),
			OriginDenomination:          row[9],
			MembershipCourseCompleted:   membershipCourseCompleted,
			MembershipCourseCompletedAt: membershipCourseCompletedAt,
			Contacted:                   contacted,
			ContactedAt:                 contactedAt,
			CreatedBy:                   userID,
			CreatedAt:                   time.Now(),
			UpdatedAt:                   time.Now(),
		}

		// minimum validation
		if member.Name == "" {
			result.Failed++
			result.Errors = append(
				result.Errors,
				"name is required",
			)
			continue
		}

		if member.Status == "" {
			member.Status = Active
		}

		if member.Status != Active && member.Status != Inactive {
			result.Failed++
			result.Errors = append(
				result.Errors,
				"invalid status for member: "+member.Name,
			)
			continue
		}

		members = append(members, member)
	}

	if len(members) == 0 {
		return result, nil
	}

	err = s.repo.CreateMany(members)
	if err != nil {
		return result, err
	}

	result.Imported = len(members)

	return result, nil
}

func parseOptionalDate(value string) *time.Time {
	if value == "" {
		return nil
	}

	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}

	return &parsed
}

func parseOptionalBool(value string) bool {
	if value == "" {
		return false
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false
	}

	return parsed
}
