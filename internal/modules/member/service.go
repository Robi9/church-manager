package member

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNameRequired  = errors.New("name is required")
	ErrInvalidStatus = errors.New("invalid status")
)

type ServiceRepository interface {
	CandidateFinder
	RunInTransaction(func(TransactionRepository) error) error
	Find(filters map[string]string, limit, offset int) ([]Member, error)
	Count(filters map[string]string) (int, error)
	GetByID(id int64) (Member, error)
	SoftDelete(id int64) error
}

type DuplicateConflictError struct {
	Result DuplicateCheckResult
}

func (e *DuplicateConflictError) Error() string {
	return "possible duplicate member found"
}

type Service struct {
	repo    ServiceRepository
	checker *DuplicateChecker
}

func NewService(repo ServiceRepository) *Service {
	return &Service{
		repo:    repo,
		checker: NewDuplicateChecker(repo),
	}
}

func (s *Service) CheckDuplicates(member Member, excludeID int64) (DuplicateCheckResult, error) {
	return s.checker.Check(member, excludeID)
}

func (s *Service) Create(member Member, forceCreate bool) (Member, error) {
	if err := prepareNewMember(&member); err != nil {
		return Member{}, err
	}

	var created Member
	err := s.repo.RunInTransaction(func(repo TransactionRepository) error {
		normalized := NormalizeMember(member)
		if err := repo.AcquireDuplicateLock(normalized.Name); err != nil {
			return err
		}

		duplicates, err := NewDuplicateChecker(repo).Check(member, 0)
		if err != nil {
			return err
		}
		if duplicates.HighestRisk == RiskHigh && !forceCreate {
			return &DuplicateConflictError{Result: duplicates}
		}

		created, err = repo.Create(member)
		if err != nil {
			return err
		}
		if forceCreate && len(duplicates.Candidates) > 0 {
			return repo.CreateDuplicateAudits(
				buildDuplicateAudits(created.ID, member.CreatedBy, "create", duplicates.Candidates),
			)
		}
		return nil
	})
	if err != nil {
		return Member{}, err
	}
	return created, nil
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

	return data, PaginationMeta{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: (total + limit - 1) / limit,
	}, nil
}

func (s *Service) Update(id int64, payload Member, forceCreate bool, userID int64) (Member, error) {
	var updated Member
	err := s.repo.RunInTransaction(func(repo TransactionRepository) error {
		current, err := repo.GetByID(id)
		if err != nil {
			return errors.New("member not found")
		}

		payload.ID = id
		payload.CreatedAt = current.CreatedAt
		payload.CreatedBy = current.CreatedBy
		payload.UpdatedAt = time.Now()
		if err := validateMember(payload); err != nil {
			return err
		}

		normalized := NormalizeMember(payload)
		if err := repo.AcquireDuplicateLock(normalized.Name); err != nil {
			return err
		}
		duplicates, err := NewDuplicateChecker(repo).Check(payload, id)
		if err != nil {
			return err
		}
		if duplicates.HighestRisk == RiskHigh && !forceCreate {
			return &DuplicateConflictError{Result: duplicates}
		}

		updated, err = repo.Update(payload)
		if err != nil {
			return err
		}
		if forceCreate && len(duplicates.Candidates) > 0 {
			return repo.CreateDuplicateAudits(
				buildDuplicateAudits(updated.ID, userID, "update", duplicates.Candidates),
			)
		}
		return nil
	})
	if err != nil {
		return Member{}, err
	}
	return updated, nil
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

type parsedImportMember struct {
	rowNumber int
	row       []string
	member    Member
}

func (s *Service) ImportCSV(file io.Reader, userID int64) (ImportResult, error) {
	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return ImportResult{}, err
	}

	result := ImportResult{}
	var headers []string
	if len(rows) > 0 {
		headers = rows[0]
	}

	parsed := make([]parsedImportMember, 0, max(0, len(rows)-1))
	now := time.Now()
	for index, row := range rows {
		if index == 0 {
			continue
		}
		member, validationError := parseImportMember(row, userID, now)
		if validationError != "" {
			addImportError(&result, index+1, row, validationError)
			continue
		}
		parsed = append(parsed, parsedImportMember{rowNumber: index + 1, row: row, member: member})
	}

	if len(parsed) > 0 {
		err = s.repo.RunInTransaction(func(repo TransactionRepository) error {
			lockNames := make([]string, 0, len(parsed))
			seenLockNames := make(map[string]struct{}, len(parsed))
			for _, item := range parsed {
				name := NormalizeName(item.member.Name)
				if name == "" {
					continue
				}
				if _, exists := seenLockNames[name]; exists {
					continue
				}
				seenLockNames[name] = struct{}{}
				lockNames = append(lockNames, name)
			}
			sort.Strings(lockNames)
			for _, name := range lockNames {
				if err := repo.AcquireDuplicateLock(name); err != nil {
					return err
				}
			}

			checker := NewDuplicateChecker(repo)
			for _, item := range parsed {
				duplicates, err := checker.Check(item.member, 0)
				if err != nil {
					return err
				}
				if duplicates.HighestRisk == RiskHigh {
					addImportError(
						&result,
						item.rowNumber,
						item.row,
						formatImportDuplicateError(duplicates.Candidates),
					)
					continue
				}
				if _, err := repo.Create(item.member); err != nil {
					return err
				}
				result.Imported++
			}
			return nil
		})
		if err != nil {
			return result, err
		}
	}

	if result.Failed > 0 {
		jobID := uuid.NewString()
		SaveImportJob(ImportJob{
			ID:        jobID,
			CreatedAt: time.Now(),
			Headers:   headers,
			ErrorRows: result.Errors,
		})
		result.JobID = jobID
	}

	return result, nil
}

func (s *Service) GetByID(id int64) (Member, error) {
	member, err := s.repo.GetByID(id)
	if err != nil {
		return Member{}, errors.New("member not found")
	}
	return member, nil
}

func prepareNewMember(member *Member) error {
	if member.Status == "" {
		member.Status = Active
	}
	if err := validateMember(*member); err != nil {
		return err
	}
	now := time.Now()
	member.CreatedAt = now
	member.UpdatedAt = now
	return nil
}

func validateMember(member Member) error {
	if NormalizeName(member.Name) == "" {
		return ErrNameRequired
	}
	if member.Status != Active && member.Status != Inactive {
		return ErrInvalidStatus
	}
	return nil
}

func buildDuplicateAudits(memberID, userID int64, operation string, candidates []DuplicateCandidate) []DuplicateAudit {
	audits := make([]DuplicateAudit, 0, len(candidates))
	for _, candidate := range candidates {
		audits = append(audits, DuplicateAudit{
			MemberID:          memberID,
			CandidateMemberID: candidate.MemberID,
			ConfirmedBy:       userID,
			Score:             candidate.Score,
			MatchedFields:     candidate.MatchedFields,
			Operation:         operation,
		})
	}
	return audits
}

func parseImportMember(row []string, userID int64, now time.Time) (Member, string) {
	if len(row) < 20 {
		return Member{}, "estrutura da linha inválida"
	}

	maritalStatus := normalizeMaritalStatus(row[7])
	if !isValidMaritalStatus(maritalStatus) {
		return Member{}, "estado civil inválido: " + row[7]
	}
	congregation, validCongregation := normalizeCongregation(row[8])
	if !validCongregation {
		return Member{}, "congregação inválida: " + row[8]
	}

	member := Member{
		Name:                        row[0],
		Phone:                       row[1],
		Status:                      normalizeStatus(row[2]),
		MemberSince:                 parseOptionalDate(row[3]),
		Baptized:                    parseOptionalBool(row[4]),
		BaptismDate:                 parseOptionalDate(row[5]),
		ChurchRole:                  row[6],
		MaritalStatus:               maritalStatus,
		Congregation:                congregation,
		OriginDenomination:          row[9],
		MembershipCourseCompleted:   parseOptionalBool(row[10]),
		MembershipCourseCompletedAt: parseOptionalDate(row[11]),
		Contacted:                   parseOptionalBool(row[12]),
		ContactedAt:                 parseOptionalDate(row[13]),
		Address:                     row[14],
		AddressNumber:               row[15],
		AddressComplement:           row[16],
		Neighborhood:                row[17],
		City:                        row[18],
		State:                       row[19],
		CreatedBy:                   userID,
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}
	if member.Status == "" {
		member.Status = Active
	}
	if err := validateMember(member); err != nil {
		switch {
		case errors.Is(err, ErrNameRequired):
			return Member{}, "nome é obrigatório"
		default:
			return Member{}, "status inválido para o membro: " + member.Name
		}
	}
	return member, ""
}

func formatImportDuplicateError(candidates []DuplicateCandidate) string {
	items := make([]string, 0)
	for _, candidate := range candidates {
		if candidate.Risk != RiskHigh {
			continue
		}
		items = append(items, fmt.Sprintf("%s (ID %d, score %d)", candidate.Name, candidate.MemberID, candidate.Score))
	}
	return "possível membro duplicado com alta probabilidade: " + strings.Join(items, "; ")
}

func parseOptionalDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "02/01/2006"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return &parsed
		}
	}
	return nil
}

func parseOptionalBool(value string) bool {
	if value == "" {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "sim" {
		return true
	}
	if normalized == "não" || normalized == "nao" {
		return false
	}
	parsed, err := strconv.ParseBool(normalized)
	return err == nil && parsed
}

func normalizeStatus(value string) Status {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "ativo", "active":
		return Active
	case "inativo", "inactive":
		return Inactive
	default:
		return Status(value)
	}
}

func normalizeMaritalStatus(value string) MaritalStatus {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "solteiro", "solteira", "single":
		return Single
	case "casado", "casada", "married":
		return Married
	case "divorciado", "divorciada", "divorced":
		return Divorced
	case "viúvo", "viúva", "viuvo", "viuva", "widowed":
		return Widowed
	default:
		return MaritalStatus(value)
	}
}

func normalizeCongregation(value string) (string, bool) {
	congregations := []string{
		"Sede", "Várzea de Cima", "Caraúno", "Cohab", "Quixadá",
		"Fortaleza", "Castelo", "Conjunto Esperança",
	}
	value = strings.TrimSpace(value)
	for _, congregation := range congregations {
		if strings.EqualFold(value, congregation) {
			return congregation, true
		}
	}
	return value, false
}

func addImportError(result *ImportResult, rowNumber int, row []string, message string) {
	result.Failed++
	result.Errors = append(result.Errors, ImportError{Row: rowNumber, Error: message, Data: row})
}

func isValidMaritalStatus(status MaritalStatus) bool {
	switch status {
	case Single, Married, Divorced, Widowed, Solteiro, Casado, Divorciado, Viúvo:
		return true
	default:
		return false
	}
}
